// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package licensing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/QuesmaOrg/quesma/platform/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"os"
)

type LicenseModule struct {
	InstallationID string
	LicenseKey     []byte
	License        *License
	Config         *config.QuesmaConfiguration
}

const (
	installationIdFile     = ".installation_id"
	quesmaAirGapModeEnvVar = "QUESMA_AIRGAP_KEY"
)

func isAirgapKeyValid(key string) bool {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))
	return keyHash == "78b14371a310f4e4f7e6c19a444f771bbe5d2c4f2154715191334bcf58420435"
}

func Init(config *config.QuesmaConfiguration) *LicenseModule {
	l := &LicenseModule{
		Config:     config,
		LicenseKey: []byte(config.LicenseKey),
	}
	// todo just for Doris to pass the license verification
	if config.ClickHouse.ConnectorType == "doris" {
		l.License = &License{
			InstallationID: "air-gapped-installation-id",
			ClientID:       "air-gapped-client-id",
		}
		return l
	}
	if airgapKey, isSet := os.LookupEnv(quesmaAirGapModeEnvVar); isSet {
		if isAirgapKeyValid(airgapKey) {
			l.logInfo("Running Quesma in airgapped mode")
			l.License = &License{
				InstallationID: "air-gapped-installation-id",
				ClientID:       "air-gapped-client-id",
			}
			return l
		}
	}
	l.logInfo("Initializing license module")
	l.Run()
	return l
}

func (l *LicenseModule) Run() {
	l.logInfo("Skip LicenseKey check")
}

func (l *LicenseModule) validateConfig() error {

	return nil
}

func (l *LicenseModule) setInstallationID() {
	if l.Config.InstallationId != "" {
		l.logInfo("Installation ID provided in the configuration [%s]", l.Config.InstallationId)
		l.InstallationID = l.Config.InstallationId
		return
	}

	if data, err := os.ReadFile(installationIdFile); err != nil {
		l.logDebug("Reading Installation ID failed [%v], generating new one", err)
		generatedID := uuid.New().String()
		l.logDebug("Generated Installation ID of [%s]", generatedID)
		l.tryStoringInstallationIdInFile(generatedID)
		l.InstallationID = generatedID
	} else {
		installationID := string(data)
		l.logDebug("Installation ID found in file [%s]", installationID)
		l.InstallationID = installationID
	}
}

func (l *LicenseModule) tryStoringInstallationIdInFile(installationID string) {
	if err := os.WriteFile(installationIdFile, []byte(installationID), 0644); err != nil {
		l.logDebug("Failed to store Installation ID in file: %v", err)
	} else {
		l.logDebug("Stored Installation ID in file [%s]", installationIdFile)
	}
}

func (l *LicenseModule) logInfo(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func (l *LicenseModule) logDebug(msg string, args ...interface{}) {
	if *l.Config.Logging.Level == zerolog.DebugLevel {
		fmt.Printf(msg+"\n", args...)
	}
}
