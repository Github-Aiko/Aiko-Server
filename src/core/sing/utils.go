package sing

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
	"github.com/inazumav/sing-box/option"
	dns "github.com/sagernet/sing-dns"
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

func getDomainStrategy(c conf.SingOptions) (dns.DomainStrategy, error) {
	switch c.DomainStrategy {
	case "":
		return dns.DomainStrategyAsIS, nil
	case "prefer_ipv4":
		return dns.DomainStrategyPreferIPv4, nil
	case "prefer_ipv6":
		return dns.DomainStrategyPreferIPv6, nil
	case "ipv4_only":
		return dns.DomainStrategyUseIPv4, nil
	case "ipv6_only":
		return dns.DomainStrategyUseIPv6, nil
	default:
		return 0, errors.New(fmt.Sprintf("unknown domain strategy: %s", c.DomainStrategy))
	}
}
