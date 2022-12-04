package test

import (
	"metaserver/config"
	"metaserver/internal/app"
	"testing"
)

func TestRunMeta1(t *testing.T) {
	var cfg *config.Config
	cfg = config.ReadConfigFrom("../../../test_conf/meta-server-1.yaml")
	app.Run(cfg)
}

func TestRunMeta2(t *testing.T) {
	var cfg *config.Config
	cfg = config.ReadConfigFrom("../../../test_conf/meta-server-2.yaml")
	app.Run(cfg)
}

func TestRunMeta3(t *testing.T) {
	var cfg *config.Config
	cfg = config.ReadConfigFrom("../../../test_conf/meta-server-3.yaml")
	app.Run(cfg)
}

func TestRunMeta4(t *testing.T) {
	var cfg *config.Config
	cfg = config.ReadConfigFrom("../../../test_conf/meta-server-4.yaml")
	app.Run(cfg)
}
