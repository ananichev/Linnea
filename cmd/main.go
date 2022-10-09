package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JoachimFlottorp/GoCommon/assert"
	"github.com/JoachimFlottorp/GoCommon/helper/json"
	"github.com/JoachimFlottorp/GoCommon/log"
	"github.com/JoachimFlottorp/Linnea/internal/config"
	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/redis"
	"github.com/JoachimFlottorp/Linnea/internal/s3"
	"github.com/JoachimFlottorp/Linnea/internal/web"
	"go.uber.org/zap"
)

var (
	cfg = flag.String("config", "config.json", "Path to the config file")
	debug = flag.Bool("debug", false, "Enable debug logging")
)

func init() {
	flag.Parse()
	
	if *debug {
		log.InitLogger(zap.NewDevelopmentConfig())
	} else {
		b := zap.NewProductionConfig()
		b.Encoding = "console"
		b.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		
		log.InitLogger(b)
	}

	if cfg == nil {
		zap.S().Fatal("Config file is not set")
	}
}

func main() {
	cfgFile, err := os.OpenFile(*cfg, os.O_RDONLY, 0)
	assert.Error(err)

	defer func() {
		err := cfgFile.Close()
		assert.Error(err)
	}();

	conf, err := json.DeserializeStruct[config.Config](cfgFile)
	assert.Error(err, "Failed to deserialize config file")

	doneSig := make(chan os.Signal, 1)
	signal.Notify(doneSig, syscall.SIGINT, syscall.SIGTERM)
	
	gCtx, cancel := ctx.WithCancel(ctx.New(context.Background(), conf))

	{
		gCtx.Inst().Redis, err = redis.Create(gCtx, redis.Options{
			Address: conf.Redis.Address,
			Username: conf.Redis.Username,
			Password: conf.Redis.Password,
			DB: conf.Redis.Database,
		})

		assert.Error(err, "Failed to create redis instance")
	}

	{
		gCtx.Inst().Storage, err = s3.New(s3.Config{
			Region: conf.S3.Region,
			AccessToken: conf.S3.AccessToken,
			SecretKey: conf.S3.SecretKey,
			Bucket: conf.S3.Bucket,
		})

		if err != nil {
			zap.S().Fatalw("Failed to create S3 storage instance", "error", err)
		}
	}

	wg := sync.WaitGroup{}

	done := make(chan any)

	go func() {
		<-doneSig
		cancel()

		go func() {
			select {
			case <-time.After(10 * time.Second):
			case <-doneSig:
			}
			zap.S().Fatal("Forced to shutdown, because the shutdown took too long")
		}()

		zap.S().Info("Shutting down")

		wg.Wait()

		zap.S().Info("Shutdown complete")
		close(done)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := web.New(gCtx, *debug); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalw("Failed to start web server", "error", err)
		}
	}()

	zap.S().Info("Ready!")
	
	<-done

	os.Exit(0)
}
