package main

type Config struct {
	UseCli bool `yaml:"use_cli"`
	UseGrpcServer bool `yaml:"use_grpc_server"`
	UseRestApi bool `yaml:"use_rest_api"`
	DictionaryServicePort string `yaml:"dictionary_service_port"`
	DatabaseHost string `yaml:"database_host"`
	DatabasePort int `yaml:"database_port"`
	DatabaseName string `yaml:"database_name"`
	DatabaseUserName string `yaml:"database_username"`
	DatabasePassword string `yaml:"database_password"`
}