package main


import (
	"os"
	"gopkg.in/yaml.v3"
)


type SMTPConfig struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type ExternalServer struct {
	IMAPServer   string `yaml:"imap_server"`
	IMAPPort     int    `yaml:"imap_port"`
	IMAPUser     string `yaml:"imap_user"`
	IMAPPassword string `yaml:"imap_password"`
	Recipient    string `yaml:"recipient"`
}

type Config struct {
	SMTP           SMTPConfig       `yaml:"smtp"`
	ExternalServers []ExternalServer `yaml:"external_servers"`
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
