package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
)

func IsCertificateExists(thumbprint string) (bool, error) {
	manager := cades.CadesManager{}
	exists, err := manager.IsCertificateExists(thumbprint)
	slog.Debug(fmt.Sprintf("Certificate with thumbprint is exists: %v", exists))
	return exists, err
}

type DeleteCertificateResult struct {
	Certificate bool `json:"certificate"`
	Container   bool `json:"container"`
}

func DeleteCertificate(thumbprint string, uniqueContainerName string) *DeleteCertificateResult {
	result := &DeleteCertificateResult{}
	manager := cades.CadesManager{}

	output, _ := manager.DeleteCertificate(thumbprint)
	if strings.Contains(output, "[ErrorCode: 0x00000000]") {
		slog.Debug(fmt.Sprintf("Certificate with thumbprint: %s deleted", thumbprint))
		result.Certificate = true
	}

	if uniqueContainerName != "" {
		output, _ = manager.DeleteContainer(uniqueContainerName)
		if strings.Contains(output, "[ErrorCode: 0x00000000]") {
			slog.Debug(fmt.Sprintf("Container with uname: %s deleted", uniqueContainerName))
			result.Container = true
		}
	}

	return result
}

func InstallPfx(path string, password string) (*cades.InstallPfxResult, error) {
	m := cades.CadesManager{}
	result, err := m.InstallPfx(path, password)
	slog.Debug(fmt.Sprintf("Install Pfx result: %+v", result))

	return result, err
}

func LinkCertWithContainer(path, containerName string) (*cades.LinkCertResult, error) {
	m := cades.CadesManager{}
	result, err := m.LinkCertWithContainer(path, containerName)
	slog.Debug(fmt.Sprintf("Link certificate with container result: %+v", result))

	return result, err
}

type Cert struct {
	Pfx         string `csv:"pfx"`
	Certificate string `csv:"cert"`
	Password    string `csv:"password"`
}

func InstallCertificate(folderPath string, cert *Cert) {
	pfxPath := filepath.Join(folderPath, cert.Pfx)
	certPath := filepath.Join(folderPath, cert.Certificate)

	if _, err := os.Stat(pfxPath); errors.Is(err, os.ErrNotExist) {
		slog.Info(err.Error())
		slog.Error(fmt.Sprintf("Pfx файл не существует: %s", pfxPath))
		return
	}

	if _, err := os.Stat(certPath); errors.Is(err, os.ErrNotExist) {
		slog.Error(fmt.Sprintf("Cer файл не существует: %s", certPath))
		return
	}

	thumbprint, err := cades.GetCertificateThumbprintFromFile(certPath)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось получить thumbprint сертификата: %s", cert.Certificate))
		return
	}

	ok, _ := IsCertificateExists(thumbprint)
	if ok {
		slog.Info(fmt.Sprintf("Сертификат[%s] с thumbprint:%s существует в хранилище: %t", cert.Certificate, thumbprint, ok))
		return
	}

	pfxResult, err := InstallPfx(pfxPath, cert.Password)
	if err != nil {
		slog.Warn("Не удалось установить контейнер из pfx файла", cert.Pfx)
		return
	} else {
		slog.Info(fmt.Sprintf("Pfx[%s] установлен, thumbprint:%s, uname:'%s'", cert.Pfx, pfxResult.Thumbprint, pfxResult.Container))
	}

	linkResult, err := LinkCertWithContainer(certPath, pfxResult.Container)
	if err != nil || !linkResult.OK {
		DeleteCertificate(thumbprint, pfxResult.Container)
		slog.Warn("Не удалось установить сертификат в закрытый контейнер, изменения отменены", pfxResult.Container, pfxResult.Thumbprint)
		return
	} else {
		slog.Info(fmt.Sprintf("Сертификат[%s] установлен в закрытый контейнер uname:'%s'", cert.Certificate, pfxResult.Container))
	}
}
