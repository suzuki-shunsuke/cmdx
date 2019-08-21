package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_setupApp(t *testing.T) {
	app := cli.NewApp()
	setupApp(app)
	assert.Equal(t, "cmdx", app.Name)
	assert.Equal(t, appUsage, app.Usage)
	assert.Equal(t, domain.Version, app.Version)
	assert.Equal(t, []cli.Author{
		{
			Name: "Shunsuke Suzuki",
		},
	}, app.Authors)
}
