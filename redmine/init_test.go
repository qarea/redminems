package redmine

import (
	"os"
	"testing"

	"github.com/powerman/narada-go/narada/staging"
)

func TestMain(m *testing.M) { os.Exit(staging.TearDown(m.Run())) }
