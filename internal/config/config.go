package config

const (
	SCHEDULER_LISTEN_PORT   = ":80"
	SCHEDULER_LISTEN_DOMAIN = "localhost"
)

const (
	Ping            = "ping"
	AuthOkAck       = "AuthOkAck"
	AuthTokenForDev = "AuthTokenForDev"
	GoodBye         = "GoodBye"
)

const (
	RepoMySQLDsn = "root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local"
	RepoLogLevel = 4
	RepoSlowMs   = 200
)

const (
	AMQP_URL  = "amqp://guest:guest@192.168.31.7:5672"
	AMQP_EXCH = "amq.direct"
)

const (
	SCHED_HEARTBEAT_INTERVAL = 5
)
