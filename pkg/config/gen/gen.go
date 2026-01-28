package main

import (
	cfg "github.com/conductorone/baton-fastly/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("fastly", cfg.Config)
}
