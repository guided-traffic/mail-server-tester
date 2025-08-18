package main


import (
	"os"
	"gopkg.in/yaml.v3"
)


type ServerConfig struct {
	Name            string `yaml:"name"`
	SMTPServer      string `yaml:"smtp_server"`
	SMTPPort        int    `yaml:"smtp_port"`
	SMTPUser        string `yaml:"smtp_user"`
	SMTPPassword    string `yaml:"smtp_password"`
	IMAPServer      string `yaml:"imap_server"`
	IMAPPort        int    `yaml:"imap_port"`
	IMAPUser        string `yaml:"imap_user"`
	IMAPPassword    string `yaml:"imap_password"`
	TLS             bool   `yaml:"tls"`
	SkipCertVerify  bool   `yaml:"skip_cert_verify"`
}

type Config struct {
	TestServer      ServerConfig   `yaml:"testserver"`
	ExternalServers []ServerConfig `yaml:"external_servers"`
	IntervalMinutes int            `yaml:"interval_minutes"`
}


func LoadConfig(path string) (*Config, error) {
       file, err := os.ReadFile(path)
       if err != nil {
	       return nil, err
       }
       var config Config
       if err := yaml.Unmarshal(file, &config); err != nil {
	       return nil, err
       }
       return &config, nil
}
