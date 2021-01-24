# 基于 MySQL 分布式锁，防止多副本应用初始化数据重复

现在有一个需求，应用启动时需要初始化一些数据，为了保证高可用，会启动多副本（replicas >= 3），如何保证数据不会重复？

## 方案一：数据带上主键
最简单的方法，初始化数据都带上主键，这样主键冲突就会报错。但是这么做我们需要对冲突的错误进行额外处理，因为插入我们一般会复用已写好的 DAO 层代码。

另外，初始化数据的主键可能是动态生成的，并不想把主键写死。所以下面来介绍此次的主角：基于 MySQL 的分布式锁的解决方案。

## 方案二：基于 MySQL 的分布式锁

多副本分布式应用，在这种 n 选 1 竞争某个资源或执行权的场景，一般都会用到分布式锁。分布式有很多种实现方式，如基于 redis，etcd，zookeeper，file 等系统。本质上，就是找个多个节点都认可的地方保存数据，通过数据竞态来实现锁，当然这个依赖最好是高可用，否则会引发单点故障。

多个副本都使用同一个 MySQL，所以我们可以很方便的基于 MySQL 实现一个分布式锁。原理很简单，利用唯一索引保证只有一个副本能插入某条数据，插入成功则表示取锁成功，执行完毕则删除该条数据释放锁。

建一个表用来存放锁数据，将 Action 设为唯一索引，表示对某个动作加锁，如：init 初始化，cronjob 定时任务等不同动作之间加锁互不影响。

```go
type lock struct {
    Id        string `gorm:"primary_key"`
    CreatedAt time.Time
    UpdatedAt time.Time
    ExpiredAt time.Time // 锁过期时间
    Action    string `gorm:"unique;not null"`
    Holder    string // 持锁人信息，可以使用 hostname
}
```

既然有过期时间，那么持锁时间设为多长合适呢？设置太短可能逻辑还没执行完锁就过期了；设置太长如果程序中途挂了没有释放锁，那么这段时间所有节点都拿不到锁。

要解决这个问题我们可以使用**租约机制（lease）**，设置较短的持锁时间，然后在持锁周期内，不断延长持锁时间，直到主动释放。这样即使程序崩溃没有 UnLock，锁也会因为没有刷新租约很快过期，不影响其他节点获取锁。

Lock 时启动一个 goroutine 刷新租约，Unlock 时通过 stopCh 将其停止。

另外，MySQL 中并没有线程去处理过期的记录，所以我们在调用 Lock 时先尝试将过期记录删掉。

核心代码：
```go
func NewLockDb(action, holder string, lease time.Duration) *lockDb {
	return &lockDb{
		db:       GetDB(context.Background()),
		stopCh:   make(chan struct{}),
		action:   action,
		holder:   holder,
		leaseAge: lease,
	}
}

func (s *lockDb) Lock() (bool, error) {
	err := s.cleanExpired()
	if err != nil {
		return false, errx.WithStackOnce(err)
	}

	err = s.db.Create(&lock{
		ExpiredAt: time.Now().Add(s.leaseAge),
		Action:    s.action,
		Holder:    s.holder,
	}).Error
	if err != nil {
		// Duplicate entry '<action_val>' for key 'action'
		if strings.Contains(err.Error(), "Duplicate entry") {
			return false, nil
		}
		return false, errx.WithStackOnce(err)
	}

	s.startLease()

	log.Debugf("%s get lock", s.holder)

	return true, nil
}

func (s *lockDb) UnLock() error {
	s.stopLease()
	var err error

	defer func() {
		err = s.db.
			Where("action = ? and holder = ?", s.action, s.holder).
			Delete(&lock{}).
			Error
	}()

	return err
}

func (s *lockDb) cleanExpired() error {
	err := s.db.
		Where("expired_at < ?", time.Now()).
		Delete(&lock{}).
		Error

	return err
}

func (s *lockDb) startLease() {
	go func() {
		// 剩余 1/4 时刷新租约
		ticker := time.NewTicker(s.leaseAge * 3 / 4)
		for {
			select {
			case <-ticker.C:
				err := s.refreshLease()
				if err != nil {
					log.Errorf("refreash lease err: %s", err)
				} else {
					log.Debug("lease refreshed")
				}
			case <-s.stopCh:
				log.Debug("lease stopped")
				return
			}
		}
	}()
}

func (s *lockDb) stopLease() {
	close(s.stopCh)
}

func (s *lockDb) refreshLease() error {
	err := s.db.Model(&lock{}).
		Where("action = ? and holder = ?", s.action, s.holder).
		Update("expired_at", time.Now().Add(s.leaseAge)).
		Error

	return err
}
```

使用及测试：
```go
func TestLock(t *testing.T) {
	i := 3
	wg := &sync.WaitGroup{}
	wg.Add(i)

	for i > 0 {
		holder := strconv.Itoa(i)
		action := "test"

		i--
		go func() {
			defer wg.Done()

			locker := dbcore.NewLockDb(action, holder, 10*time.Second)

			if _, err := locker.Lock(); err != nil {
				t.Logf("not hold the lock, err: %+v", err)
				return
			}

			time.Sleep(30 * time.Second)
			locker.UnLock()
		}()
	}

	wg.Wait()
}
```

完整代码：https://github.com/win5do/go-microservice-demo/blob/main/pkg/repository/db/dbcore/lock.go

这个分布式锁实现在初始数据场景是够用了，但并不完美，例如：依赖时间同步，不能容忍**时间偏斜**；获取锁不是阻塞的，如果要抢锁需要使用方自旋； 锁不可重入，粒度是进程级别，同一个 Action，当前进程获取锁后，释放后才能再次获取锁。

大家可以思考一下如何完善。