package data

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"

	"github.com/toomanysource/atreus/app/comment/service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ProviderSet = wire.NewSet(NewData, NewKafkaWriter, NewCommentRepo, NewUserRepo, NewMysqlConn, NewRedisConn)

type Data struct {
	db    *gorm.DB
	cache *redis.Client
	kfk   *kafka.Writer
	log   *log.Helper
}

func NewData(db *gorm.DB, cacheClient *redis.Client, kfk *kafka.Writer, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(log.With(logger, "module", "data"))
	// 并发关闭所有数据库连接，后期根据Redis与Mysql是否数据同步修改
	cleanup := func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cacheClient.Ping(context.Background()).Result()
			if err != nil {
				logHelper.Warn("Redis connection pool is empty")
				return
			}
			if err = cacheClient.Close(); err != nil {
				logHelper.Errorf("Redis connection closure failed, err: %w", err)
			}
			logHelper.Info("Successfully close the Redis connection")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := kfk.Close(); err != nil {
				logHelper.Errorf("Kafka connection closure failed, err: %w", err)
			}
			logHelper.Info("Successfully close the Kafka connection")
		}()
		wg.Wait()
	}

	data := &Data{
		db:    db.Model(&Comment{}),
		cache: cacheClient,
		kfk:   kfk,
		log:   logHelper,
	}
	return data, cleanup, nil
}

// NewMysqlConn mysql数据库连接
func NewMysqlConn(c *conf.Data, l log.Logger) *gorm.DB {
	logs := log.NewHelper(log.With(l, "module", "data/mysql"))
	db, err := gorm.Open(mysql.Open(c.Mysql.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logs.Fatalf("Database connection failure, err : %v", err)
	}
	InitDB(db)
	logs.Info("Database enabled successfully!")
	return db.Model(&Comment{})
}

// NewRedisConn Redis数据库连接
func NewRedisConn(c *conf.Data, l log.Logger) *redis.Client {
	logs := log.NewHelper(log.With(l, "module", "data/redis"))
	client := redis.NewClient(&redis.Options{
		DB:           int(c.Redis.CommentDb),
		Addr:         c.Redis.Addr,
		Username:     c.Redis.Username,
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})

	// ping Redis客户端，判断连接是否存在
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logs.Fatalf("Redis database connection failure, err : %v", err)
	}
	logs.Info("Cache enabled successfully!")
	return client
}

func NewKafkaWriter(c *conf.Data) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(c.Kafka.Addr),
		Topic:                  c.Kafka.Topic,
		Balancer:               &kafka.LeastBytes{},
		WriteTimeout:           c.Kafka.WriteTimeout.AsDuration(),
		ReadTimeout:            c.Kafka.ReadTimeout.AsDuration(),
		AllowAutoTopicCreation: true,
	}
	log.Info("Kafka enabled successfully!")
	return writer
}

// InitDB 创建Comments数据表，并自动迁移
func InitDB(db *gorm.DB) {
	if err := db.AutoMigrate(&Comment{}); err != nil {
		log.Fatalf("Database initialization error, err : %v", err)
	}
}
