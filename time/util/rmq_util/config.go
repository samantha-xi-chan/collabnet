package rmq_util

// AMQP is the amqp configuration
type AMQP struct {
	URL      string `default:"amqp://guest:guest@192.168.31.6:5672/guest"`
	Exchange string `default:"amq.direct"`
}

// Config is the application configuration
type Config struct {
	//AppName    string `json:"app-name" default:"rabbitmq"`
	//AppVersion string `json:"app-version" required:"true"`

	AMQP AMQP
}
