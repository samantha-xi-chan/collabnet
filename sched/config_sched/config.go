package config_sched

const (
	TESTCASE_CNT = 1

	//TEST_TIME_PREPARE    = 5
	//TEST_TIMEOUT_PREPARE = TEST_TIME_PREPARE * 20

	//TEST_TIME_RUN    = 10
	//TEST_TIMEOUT_RUN = TEST_TIME_RUN * 20
)

const (
	SCHED_HEARTBEAT_INTERVAL = 30 /* Second */
	SCHED_HEARTBEAT_TIMEOUT  = 60 /* Second */

	CMD_ACK_TIMEOUT = 10 /* Second */
)

const (
	SCHEDULER_LISTEN_PORT = ":80"
	//SCHEDULER_LISTEN_DOMAIN = "localhost"
)

const (
	Ping            = "ping"
	AuthOkAck       = "AuthOkAck"
	AuthTokenForDev = "AuthTokenForDev"
	GoodBye         = "GoodBye"
)

const (
	//RepoMySQLDsn = "root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local"
	RepoLogLevel = 2
	RepoSlowMs   = 200
)

const (
	AMQP_EXCH = "amq.direct"
)

const ()
