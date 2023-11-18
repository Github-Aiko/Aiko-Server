package cmd

import (
	"crypto/x509"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/protocol/tls/cert"
)

var certCommand = &cobra.Command{
	Use:   "cert",
	Short: "Generate TLS certificates",
	Run:   executeCert,
}

var (
	certDomainNames  []string
	certCommonName   string
	certOrganization string
	certIsCA         bool
	certExpireDays   int
	certFilePath     string
)

func init() {
	command.AddCommand(certCommand)
	certCommand.Flags().StringArrayVarP(&certDomainNames, "domain", "", nil, "Domain name for the certificate")
	certCommand.Flags().StringVarP(&certCommonName, "name", "", "Aiko-Server Inc", "The common name of this certificate")
	certCommand.Flags().StringVarP(&certOrganization, "org", "", "Aiko-Server Inc", "Organization of the certificate")
	certCommand.Flags().BoolVarP(&certIsCA, "ca", "", false, "Whether this certificate is a CA")
	certCommand.Flags().IntVarP(&certExpireDays, "expire", "", 90, "Number of days until the certificate expires. Default value 90 days.")
	certCommand.Flags().StringVarP(&certFilePath, "output", "", "/etc/Aiko-Server/cert", "Save certificate in file.")
}

func executeCert(cmd *cobra.Command, args []string) {
	var opts []cert.Option
	if certIsCA {
		opts = append(opts, cert.Authority(certIsCA))
		opts = append(opts, cert.KeyUsage(x509.KeyUsageCertSign|x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature))
	}

	opts = append(opts, cert.NotAfter(time.Now().Add(time.Duration(certExpireDays)*24*time.Hour)))
	opts = append(opts, cert.CommonName(certCommonName))
	if len(certDomainNames) > 0 {
		opts = append(opts, cert.DNSNames(certDomainNames...))
	}
	opts = append(opts, cert.Organization(certOrganization))

	certificate, err := cert.Generate(nil, opts...)
	if err != nil {
		log.Fatalf("failed to generate TLS certificate: %s", err)
	}

	if certFilePath != "" {
		cert_FilePath := certFilePath + "/aiko_server.cert"
		key_FilePath := certFilePath + "/aiko_server.key"

		if err := createDirectoryIfNotExists(certFilePath); err != nil {
			log.Fatalf("failed to create directory: %s", err)
		}

		if err := saveCertificateAndKey(certificate, cert_FilePath, key_FilePath); err != nil {
			log.Fatalf("failed to save files: %s", err)
		}
	}
}

func createDirectoryIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func saveCertificateAndKey(certificate *cert.Certificate, certFilePath, keyFilePath string) error {
	certPEM, keyPEM := certificate.ToPEM()

	if err := writeFile(certPEM, certFilePath); err != nil {
		return err
	}

	if err := writeFile(keyPEM, keyFilePath); err != nil {
		return err
	}

	return nil
}

func writeFile(content []byte, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return common.Error2(f.Write(content))
}
