package file

import (
	"strings"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/config"
	"github.com/sado0823/go-kitx/internal/test/pbconfig"

	"github.com/stretchr/testify/assert"
)

func Test_New_Fail(t *testing.T) {
	t.Run("with dir", func(t *testing.T) {
		assert.NotPanics(t, func() {
			f := New("./")
			err := f.Load()
			assert.NotNil(t, err)
			t.Log(err)

			assert.NotPanics(t, func() {
				mp := make(map[string]interface{})
				err := f.Scan(&mp)
				assert.NotNil(t, err)
				t.Log(err)
			})
		})
	})
}

type selfBoot struct {
	Server *struct {
		GoRoot string `json:"go_root" yaml:"goRoot"`
		Http   *struct {
			Addr    string `json:"addr"`
			Timeout string `json:"timeout"`
		} `json:"http"`
		Grpc *struct {
			Addr    string `json:"addr"`
			Timeout string `json:"timeout"`
		} `json:"grpc"`
	}
	Data *struct {
		Database *struct {
			Driver string `json:"driver"`
			Source string `json:"source"`
		} `json:"database"`
		Redis *struct {
			Addr         string `json:"addr"`
			ReadTimeout  string `json:"read_timeout" yaml:"readTimeout"`
			WriteTimeout string `json:"write_timeout" yaml:"writeTimeout"`
		} `json:"redis"`
	}
}

func Test_New_No_ENV(t *testing.T) {
	cases := []string{"./test.json", "./test.toml", "./test.yaml"}
	for _, caseV := range cases {
		t.Run(caseV, func(t *testing.T) {
			configer := config.New(config.WithReader(New(caseV)))

			assert.Nil(t, configer.Load())

			t.Run("with pb struct", func(t *testing.T) {
				var mm = new(pbconfig.Bootstrap)
				assert.Nil(t, configer.Scan(mm))
				t.Log(mm)
				assertPb(t, mm, false)
			})

			t.Run("with self struct", func(t *testing.T) {
				var mm = new(selfBoot)
				assert.Nil(t, configer.Scan(mm))
				t.Logf("%#v", mm)
				assertSelf(t, mm, false)
			})
		})
	}
}

func Test_New_With_ENV(t *testing.T) {
	cases := []string{"./test.json", "./test.toml", "./test.yaml"}
	for _, caseV := range cases {
		t.Run(caseV, func(t *testing.T) {
			configer := config.New(config.WithReader(New(caseV, WithEnv())))

			assert.Nil(t, configer.Load())

			t.Run("with pb struct", func(t *testing.T) {
				var mm = new(pbconfig.Bootstrap)
				assert.Nil(t, configer.Scan(mm))
				t.Log(mm)
				assertPb(t, mm, true)
			})

			t.Run("with self struct", func(t *testing.T) {
				var mm = new(selfBoot)
				assert.Nil(t, configer.Scan(mm))
				t.Logf("%#v", mm)
				assertSelf(t, mm, true)
			})
		})
	}
}

func assertPb(t *testing.T, pb *pbconfig.Bootstrap, WithEnv bool) {
	t.Log("from env: ", pb.Server.GoRoot)
	if WithEnv {
		// from env
		assert.NotEmpty(t, pb.Server.GoRoot)
		envs := strings.Split(pb.Server.GoRoot, "/")
		assert.Equal(t, "go", envs[len(envs)-1])
	} else {
		assert.Equal(t, "${GOROOT}", pb.Server.GoRoot)
	}

	// http
	assert.Equal(t, "0.0.0.0:8000", pb.Server.Http.Addr)
	assert.Equal(t, int64(1), pb.Server.Http.Timeout.Seconds)

	// grpc
	assert.Equal(t, "0.0.0.0:9000", pb.Server.Grpc.Addr)
	assert.Equal(t, int64(1), pb.Server.Grpc.Timeout.Seconds)

	// mysql
	assert.Equal(t, "mysql", pb.Data.Database.Driver)
	assert.Equal(t, "root:root@tcp(127.0.0.1:3306)/test", pb.Data.Database.Source)

	// redis
	assert.Equal(t, "127.0.0.1:6379", pb.Data.Redis.Addr)
	assert.Equal(t, time.Duration(200000000), pb.Data.Redis.ReadTimeout.AsDuration())
	assert.Equal(t, time.Duration(200000000), pb.Data.Redis.WriteTimeout.AsDuration())

}

func assertSelf(t *testing.T, self *selfBoot, WithEnv bool) {
	t.Log("from env: ", self.Server.GoRoot)
	if WithEnv {
		// from env
		assert.NotEmpty(t, self.Server.GoRoot)
		envs := strings.Split(self.Server.GoRoot, "/")
		assert.Equal(t, "go", envs[len(envs)-1])
	} else {
		assert.Equal(t, "${GOROOT}", self.Server.GoRoot)
	}

	// http
	assert.Equal(t, "0.0.0.0:8000", self.Server.Http.Addr)
	assert.Equal(t, "1s", self.Server.Http.Timeout)

	// grpc
	assert.Equal(t, "0.0.0.0:9000", self.Server.Grpc.Addr)
	assert.Equal(t, "1s", self.Server.Grpc.Timeout)

	// mysql
	assert.Equal(t, "mysql", self.Data.Database.Driver)
	assert.Equal(t, "root:root@tcp(127.0.0.1:3306)/test", self.Data.Database.Source)

	// redis
	assert.Equal(t, "127.0.0.1:6379", self.Data.Redis.Addr)
	assert.Equal(t, "0.2s", self.Data.Redis.ReadTimeout)
	assert.Equal(t, "0.2s", self.Data.Redis.WriteTimeout)

}
