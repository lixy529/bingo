// 项目相关配置
//   变更历史
//     2017-02-09  lixiaoya  新建
package bingo

import (
	"errors"
	"fmt"
	"legitlab.letv.cn/uc_tp/goweb/config"
	"os"
	"path"
	"strings"
	"time"
)

const (
	DbDef    = "db"
	DbPre    = "db_"
	CacheDef = "cache"
	CachePre = "cache_"
	MongoDef = "mongo"
	MongoPre = "mongo_"
	MqDef    = "mq"
	MqPre    = "mq_"
)

var (
	GlobalCfg *config.Config
	AppCfg    *AppConfig
	AppRoot   string // 程序根目录
)

func init() {
	// 初始化配置信息
	var err error
	AppCfg, err = newAppConfig()
	if err != nil {
		panic(err.Error())
	}
}

// AppConfig app相关配置
type AppConfig struct {
	AppName   string // 应用名称
	RunMode   string // 运行模式: dev | prod
	LogName   string // 框架使用的日志名称
	LogCfg    string // 框架使用的日志配置
	ServerCfg ServerConfig
	WebCfg    WebConfig
	SessCfg   SessionConfig
	DbConfigs []DbConfig
	CacheCfgs []CacheConfig
	MongoCfgs []MongoConfig
	Log       LogConfig  // 业务使用的日志配置
	LangCfg   LangConfig // 语言包配置
	MqConfigs map[string]*MqConfig
}

// ListenConfig 服务监听相关配置
type ServerConfig struct {
	PidFile      string
	UseGrace     bool          // 是否使用grace启动程序
	ReadTimeout  time.Duration // 读超时时间，单位秒
	WriteTimeout time.Duration // 写超时时间，单位秒
	ShutTimeout  time.Duration // shutdown服务的超时间
	GzipStatus   bool          // 压缩状态
	GzipLevel    int           // 压缩水平，取值为0-NoCompression 1-BestSpeed 9-BestCompression -1-DefaultCompression -2-HuffmanOnly
	GzipMinLen   int           // 最小压缩长度，小于0表示不压缩

	Secure     bool // true:https false:http
	IsFcgi     bool
	Addr       string
	Port       int
	ReqTimeout time.Duration // 请求超时时间，单位秒
	MaxGoCnt   int           // 最大协程数，<=0 不限制
	CertFile   string
	KeyFile    string

	ForwardName string // 有代理转发时需要设置，获取真实的客户端IP
	ForwardRev  bool   // true-按倒序排，false-按顺序排
}

// WebConfig web相关配置
type WebConfig struct {
	StaticDir []string // 静态文件目录
	ViewsDir  string   // 模板文件目录，默认views
	ViewsExt  string   // 模板文件扩展名，默认html
}

// SessionConfig session相关配置
type SessionConfig struct {
	SessOn         bool
	ProviderName   string
	ProviderConfig string
	LifeTime       int64
	CookieName     string
}

// DbConfig Db相关配置
type DbConfig struct {
	dbName     string   // db名称，比如分地区，取配置段名，db_pass=> pass | db => db | db_ => db
	driverName string   // 数据种类，如mysql
	maxOpen    int      // 最大连接数， 如果值小于等于0将不限制
	maxIdle    int      // 最大空闲连接数，如果值小于0将不做设置，等于0不保留空闲连接
	maxLife    int64    // 可被重新使用的最大时间间隔，如果小于0将永久重用，单位秒
	master     string   // 主库配置
	slaves     []string // 从库配置，可以配置多个
}

// CacheConfig Cache相关配置
type CacheConfig struct {
	cacheName   string // Cache名称
	cacheType   string // Cache类型，目前支持redis、memcache
	cacheConfig string // Cache连接串
}

// MongoConfig Mongo相关配置
type MongoConfig struct {
	mongoName string // Mongo实例名称
	connStr   string //数据库连接串
	mode      int    // 数据库读写优先模式
	/*
		Primary            Mode = 2 // Default mode. All operations read from the current replica set primary.
		PrimaryPreferred   Mode = 3 // Read from the primary if available. Read from the secondary otherwise.
		Secondary          Mode = 4 // Read from one of the nearest secondary members of the replica set.
		SecondaryPreferred Mode = 5 // Read from one of the nearest secondaries if available. Read from primary otherwise.
		Nearest            Mode = 6 // Read from one of the nearest members, irrespective of it being primary or secondary.
		Eventual           Mode = 0 // Same as Nearest, but may change servers between reads.
		Monotonic          Mode = 1 // Same as SecondaryPreferred before first write. Same as Primary after first write.
		Strong             Mode = 2 // Same as Primary.
	*/
	maxPoolSize int //连接池最大连接数，默认1024
	timeout     int //连接超时时间，默认3s
}

// SyslogConfig
type LogConfig struct {
	LogName string // 业务使用的日志名称
	LogCfg  string // 业务使用的日志配置
}

// LangConfig 语言包配置
type LangConfig struct {
	LangPath string // 语言包目录
}

type MqConfig struct {
	cnfdata map[string]interface{}
}

// newAppConfig 解析配置文件
// 配置先从参数（-c config file）取，如果没有传配置文件参数则从环境变量取（APPCONFIG），默认为app.conf
//   参数
//     void
//   返回
//     成功时返回AppConfig实例化对象，失败时返回错误信息
func newAppConfig() (*AppConfig, error) {
	// APPROOT
	AppRoot = os.Getenv("APPROOT") // 环境变量里必须添加APPROOT
	if len(AppRoot) == 0 {
		return nil, errors.New("config: Please set the 'APPROOT' environment variable")
	}

	// 配置文件
	cfgFile := os.Getenv("APPCONFIG") // 环境变量里可以配置配置文件名，未配置默认取app.conf
	if cfgFile == "" {
		cfgFile = "config/app.conf"
	}

	if !path.IsAbs(cfgFile) {
		cfgFile = path.Join(AppRoot, "config", cfgFile)
	}

	// 初始化配置文件
	var err error
	GlobalCfg, err = config.NewConfig(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("config: New confir err: %s", err.Error())
	}

	// 获取压缩信息
	gzipStatus, gzipLevel, gzipMinLen := getGzip()

	return &AppConfig{
		AppName: GlobalCfg.GetString("app", "app_name", "App"),
		RunMode: GlobalCfg.GetString("app", "run_mode", PROD),
		LogName: "frame",
		LogCfg:  fmt.Sprintf(`{"FilePath":"%s/log","filename":"goweb.log","maxlines":0,"maxsize":4000,"perm":"0660","level":1, "showcall":true, "depth":3}`, AppRoot),

		ServerCfg: ServerConfig{
			PidFile:      getPidFile(),
			UseGrace:     GlobalCfg.GetBool("server", "use_grace", false),
			ReadTimeout:  time.Duration(GlobalCfg.GetInt("server", "read_timeout", 0)),
			WriteTimeout: time.Duration(GlobalCfg.GetInt("server", "write_timeout", 0)),
			ShutTimeout:  time.Duration(GlobalCfg.GetInt("server", "shut_timeout", 0)),

			Secure:     GlobalCfg.GetBool("server", "secure", false),
			IsFcgi:     GlobalCfg.GetBool("server", "is_fcgi", false),
			Addr:       GlobalCfg.GetString("server", "addr", ""),
			Port:       GlobalCfg.GetInt("server", "port", 0),
			ReqTimeout: time.Duration(GlobalCfg.GetInt("server", "req_timeout", 10)),
			MaxGoCnt:   GlobalCfg.GetInt("server", "max_gocnt", 0),
			CertFile:   GlobalCfg.GetString("server", "cert_file", ""),
			KeyFile:    GlobalCfg.GetString("server", "key_file", ""),

			GzipStatus: gzipStatus,
			GzipLevel:  gzipLevel,
			GzipMinLen: gzipMinLen,

			ForwardName: GlobalCfg.GetString("server", "forward_name", ""),
			ForwardRev:  GlobalCfg.GetBool("server", "forward_rev", true),
		},

		WebCfg: WebConfig{
			StaticDir: strings.Split(GlobalCfg.GetString("web", "static_dir", "prod"), ","),
			ViewsDir:  getViewsDir(),
			ViewsExt:  GlobalCfg.GetString("web", "views_ext", ".html"),
		},

		SessCfg: SessionConfig{
			SessOn:         GlobalCfg.GetBool("session", "sess_on", false),
			ProviderName:   GlobalCfg.GetString("session", "provider_name", "memory"),
			ProviderConfig: GlobalCfg.GetString("session", "provider_config", ""),
			LifeTime:       GlobalCfg.GetInt64("session", "life_time", 3600),
			CookieName:     GlobalCfg.GetString("session", "cookie_name", "GOSESSIONID"),
		},
		DbConfigs: getDbConfig(),
		CacheCfgs: getCacheCfg(),
		MongoCfgs: getMongoCfg(),
		MqConfigs: getMqConfigs(),
		Log: LogConfig{
			LogName: GlobalCfg.GetString("log", "log_name", "console"),
			LogCfg:  GlobalCfg.GetString("log", "log_config", ""),
		},
		LangCfg: LangConfig{
			LangPath: getLangPath(),
		},
	}, nil
}

// getPidFile 返回进程pid文件路径
//   参数
//
//   返回
//     pid文件
func getPidFile() string {
	pidFile := GlobalCfg.GetString("server", "pid_file", "")
	if len(pidFile) == 0 {
		pidFile = GlobalCfg.GetString("app", "app_name", "App") + ".pid"
	}

	if !path.IsAbs(pidFile) {
		// 相对路径
		pidFile = path.Join(AppRoot, "log", pidFile)
	}

	return pidFile
}

// getViewsDir 返回模板的目录
//   参数
//
//   返回
//     模板目录
func getViewsDir() string {
	v := GlobalCfg.GetString("web", "views_dir", "views")
	if path.IsAbs(v) {
		return v
	}
	return path.Join(AppRoot, v)
}

// getDbConfig 获取db配置
//   参数
//
//   返回
//     数据库配置
func getDbConfig() []DbConfig {
	var dbCfgs []DbConfig
	secs := GlobalCfg.GetSecs()
	for _, sec := range secs {
		sec = strings.ToLower(sec)
		if sec == DbDef || strings.HasPrefix(sec, DbPre) {
			var cfg DbConfig
			n := len(DbPre)
			if sec == DbDef || sec == DbPre {
				cfg.dbName = DbDef
			} else {
				cfg.dbName = sec[n:]
			}

			cfg.driverName = GlobalCfg.GetString(sec, "driver_name", "mysql")
			cfg.maxOpen = GlobalCfg.GetInt(sec, "max_open", 0)
			cfg.maxIdle = GlobalCfg.GetInt(sec, "max_idle", -1)
			cfg.maxLife = GlobalCfg.GetInt64(sec, "max_life", 28800) // 默认8小时
			cfg.master = GlobalCfg.GetString(sec, "master")
			if names, ok := GlobalCfg.GetSec(sec); ok {
				for name := range names {
					name = strings.ToLower(name)
					if strings.HasPrefix(name, "slave") {
						val := GlobalCfg.GetString(sec, name)
						if val != "" {
							cfg.slaves = append(cfg.slaves, val)
						}
					}
				}
			}

			dbCfgs = append(dbCfgs, cfg)
		}
	}

	return dbCfgs
}

// getGzip 获取gzip信息
//   参数
//     void
//   返回
//     压缩状态、压缩水平、压缩最小长度
func getGzip() (bool, int, int) {
	status := true
	level := GlobalCfg.GetInt("server", "gzip_level", -1)
	if level != 0 && level != 1 && level != 9 && level != -1 && level != -2 {
		level = -1
	}

	isFcgi := GlobalCfg.GetBool("server", "is_fcgi", false)
	// fcgi不使用压缩，网页服务器自己支持，比如nginx
	if level == 0 || isFcgi {
		status = false
	}

	minLen := GlobalCfg.GetInt("server", "gzip_min", 0)
	if level <= 0 {
		level = 0
	}

	return status, level, minLen

}

// getCacheCfg 获取cache配置
//   参数
//
//   返回
//     缓存配置信息
func getCacheCfg() []CacheConfig {
	var cacheCfgs []CacheConfig
	secs := GlobalCfg.GetSecs()
	for _, sec := range secs {
		sec = strings.ToLower(sec)
		n := len(CachePre)
		if sec == CacheDef || strings.HasPrefix(sec, CachePre) {
			var cfg CacheConfig

			if sec == CacheDef || sec == CachePre {
				cfg.cacheName = CacheDef
			} else {
				cfg.cacheName = sec[n:]
			}

			cfg.cacheType = GlobalCfg.GetString(sec, "cache_type")
			cfg.cacheConfig = GlobalCfg.GetString(sec, "cache_config")
			cacheCfgs = append(cacheCfgs, cfg)
		}
	}

	return cacheCfgs
}

// getMongoCfg 获取mongo配置
//   参数
//
//   返回
//    Mongo配置信息
func getMongoCfg() []MongoConfig {
	var mongoCfgs []MongoConfig
	secs := GlobalCfg.GetSecs()
	for _, sec := range secs {
		sec = strings.ToLower(sec)
		n := len(MongoPre)
		if sec == MongoDef || strings.HasPrefix(sec, MongoPre) {
			var cfg MongoConfig
			if sec == MongoDef || sec == MongoPre {
				cfg.mongoName = MongoDef
			} else {
				cfg.mongoName = sec[n:]
			}

			cfg.connStr = GlobalCfg.GetString(sec, "conn_str")
			cfg.mode = GlobalCfg.GetInt(sec, "mode")
			cfg.maxPoolSize = GlobalCfg.GetInt(sec, "max_pool_size")
			cfg.timeout = GlobalCfg.GetInt(sec, "timeout")
			mongoCfgs = append(mongoCfgs, cfg)
		}
	}
	return mongoCfgs
}

// getLangPath 返回语言包目录
//   参数
//     void
//   返回
//     模板根目录
func getLangPath() string {
	langPath := GlobalCfg.GetString("lang", "lang_path", "")
	if !path.IsAbs(langPath) {
		langPath = path.Join(AppRoot, langPath)
	}

	return langPath
}

// GetString 根据key值获取对应的value值，返回结果为string型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func GetString(section, key string, def ...string) string {
	return GlobalCfg.GetString(section, key, def...)
}

// GetBool 根据key值获取对应的value值，返回结果为bool型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func GetBool(section, key string, def ...bool) bool {
	return GlobalCfg.GetBool(section, key, def...)
}

// GetInt 根据key值获取对应的value值，返回结果为int型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func GetInt(section, key string, def ...int) int {
	return GlobalCfg.GetInt(section, key, def...)
}

// GetInt64 根据key值获取对应的value值，返回结果为int64型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func GetInt64(section, key string, def ...int64) int64 {
	return GlobalCfg.GetInt64(section, key, def...)
}

// GetFloat64 根据key值获取对应的value值，返回结果为float64型
// 如果key值不存在就返回默认值
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//     Value值
func GetFloat64(section, key string, def ...float64) float64 {
	return GlobalCfg.GetFloat64(section, key, def...)
}

// SetValue 设置一个key值
// 如果key值不存在，则新建一个
// 如果key值存在，则更新
//   参数
//     section: 段名
//     key:     key值
//     def:     默认值
//   返回
//
func SetValue(section, key, value string) {
	GlobalCfg.SetValue(section, key, value)
}

// GetSec 根据section值获取段下所有的配置
//   参数
//     section: 段名
//   返回
//     段下对应的所有Value值
func GetSec(section string) (map[string]string, bool) {
	return GlobalCfg.GetSec(section)
}

// GetSecs 获取所有段名
//   参数
//
//   返回
//     所有段列表
func GetSecs() []string {
	return GlobalCfg.GetSecs()
}

/*
   初始化队列配置
*/
func getMqConfigs() map[string]*MqConfig {
	configs := make(map[string]*MqConfig)

	secs := GlobalCfg.GetSecs()
	var mqName string
	for _, sec := range secs {
		sec = strings.ToLower(sec)
		n := len(MqPre)
		if sec == MqDef || strings.HasPrefix(sec, MqPre) {
			if sec == MqDef || sec == MqPre {
				mqName = MqDef
			} else {
				mqName = sec[n:]
			}
			mqConfig, ret := GlobalCfg.GetSec(sec)
			if ret {
				mqCnf := make(map[string]interface{})
				for key, val := range mqConfig {
					switch key {
					case "HOST", "QUEUE", "ADAPTER":
						mqCnf[strings.ToLower(key)] = val
					default:
						mqCnf[strings.ToLower(key)] = GetInt(sec, key, 0)
					}
				}
				configs[mqName] = &MqConfig{cnfdata: mqCnf}
			}
		}
	}

	return configs
}

/*
*
 */
func (ac *AppConfig) GetMqConfig(mqName string) map[string]interface{} {
	if config, exists := ac.MqConfigs[mqName]; exists {
		return config.cnfdata
	}
	return nil
}
