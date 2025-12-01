package core

import (
	"fmt"
	"os"
	"path/filepath"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"golang.org/x/exp/slog"
)

func extractPublicKeyFromCertificates(certificatesPath []string) map[string]string {
	certificatePublicKeys := make(map[string]string)
	for _, path := range certificatesPath {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Debug(fmt.Sprintf("Cant read file %s: %v", path, err))
			continue
		}

		certificate, err := cades.LoadCertificate(data)
		if err != nil {
			slog.Debug(fmt.Sprintf("Cant parse certificate %s: %v", path, err))
			continue
		}

		if certificate.IsCA || certificate.PublicKey != nil {
			continue
		}

		subjectPublicKeyInfo, err := cades.ParseSubjectPublicKeyInfo(certificate)
		if err != nil {
			continue
		}

		shortPublicKey := cades.GetCertificateShortPublicKey(subjectPublicKeyInfo)
		certificatePublicKeys[shortPublicKey] = path
	}

	return certificatePublicKeys
}

func extractPublicKeyFromContainers(containersPath []string) map[string]string {
	containerPublicKeys := make(map[string]string)
	for _, path := range containersPath {
		headerPath := filepath.Join(path, "header.key")

		headerData, err := os.ReadFile(headerPath)
		if err != nil {
			slog.Debug(fmt.Sprintf("Cant read container header file %s: %v", headerPath, err))
			continue
		}

		publicKey := cades.GetShortPublicKeyFromPrivateKey(headerData)
		containerPublicKeys[publicKey] = path
	}
	return containerPublicKeys
}

type DigitalSignaturePair struct {
	Certificate string
	Container   string
}

func FindDigitalSignaturePairs(path string) ([]*ESignatureInstallParams, error) {
	var result []*ESignatureInstallParams
	var certificates []string
	var containers []string

	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "root" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".cer" {
			certificates = append(certificates, path)
		}

		if info.Name() == "header.key" {
			containers = append(containers, filepath.Dir(path))
		}

		return err
	})

	certificatePublicKeys := extractPublicKeyFromCertificates(certificates)
	containerPublicKeys := extractPublicKeyFromContainers(containers)
	slog.Debug(fmt.Sprintf("Certificate count: %d", len(certificatePublicKeys)))
	slog.Debug(fmt.Sprintf("Container count: %d", len(containerPublicKeys)))

	for k := range certificatePublicKeys {
		_, ok := containerPublicKeys[k]
		if !ok {
			slog.Debug(fmt.Sprintf("Pair not found for %s, %v", k, certificatePublicKeys[k]))
			continue
		}

		certificatePath, err := filepath.Rel(path, certificatePublicKeys[k])
		if err != nil {
			slog.Debug(err.Error())
			continue
		}

		containerPath, err := filepath.Rel(path, containerPublicKeys[k])
		if err != nil {
			slog.Debug(err.Error())
			continue
		}

		result = append(result, &ESignatureInstallParams{
			CertificatePath: certificatePath,
			ContainerPath:   containerPath,
		})
	}
	slog.Debug(fmt.Sprintf("Found %d pair", len(result)))

	return result, nil
}
