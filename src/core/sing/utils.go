package sing

import (
	"fmt"
	"strconv"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
	"github.com/inazumav/sing-box/option"
)

func processFallback(c *conf.Options, fallbackForALPN map[string]*option.ServerOptions) error {
	for k, v := range c.SingOptions.FallBackConfigs.FallBackForALPN {
		fallbackPort, err := strconv.Atoi(v.ServerPort)
		if err != nil {
			return fmt.Errorf("unable to parse fallbackForALPN server port error: %s", err)
		}
		fallbackForALPN[k] = &option.ServerOptions{Server: v.Server, ServerPort: uint16(fallbackPort)}
	}
	return nil
}
