package axscheduler

import (
	ax_tools "gogate/shared/axtools"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ax_tools.InitLogger("debug")
	code := m.Run()
	os.Exit(code)
}
