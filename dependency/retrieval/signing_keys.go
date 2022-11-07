package retrieval

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joshuatcasey/libdownloader"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"golang.org/x/crypto/openpgp"
)

// retrieved from https://www.php.net/gpg-keys.php on 2022-11-07
var phpSigningKeys = map[string][]string{
	"8.1": {
		"528995BFEDFBA7191D46839EF9BA0ADA31CBD89E",
		"39B641343D8C104B2B146DC3F9C39DC0B9698544",
		"F1F692238FBC1666E5A5CCD4199F9DFEF6FFBAFD",
	},
	"8.0": {
		"1729F83938DA44E27BA0F4D3DBDB397470D12172",
		"BFDDD28642824F8118EF77909B67A5C12229118F",
		//"2C16C765DBE54A088130F1BC4B9B5F600B55F3B4", // fingerprint not found on openpgp.org
	},
	"7.4": {
		"5A52880781F755608BF815FC910DEB46F53EA312",
		"42670A7FE4D0441C8E4632349E4FDC074A4EF02D",
	},
}

func getGpgUrl(fingerprint string) string {
	return fmt.Sprintf("https://keys.openpgp.org/vks/v1/by-fingerprint/%s", fingerprint)
}

// verifyChecksumAndSignature will retrieve the Public keys from openpgp.org.
// This is as safe as making a copy since an attacker would need to gain control over both php.net and openpgp.org in order
// to compromise the source download, the source signature, and the public key.
func verifyChecksumAndSignature(phpRelease PhpReleasePretty, sourceURL string, logger scribe.Emitter) error {
	version := phpRelease.Version().String()

	tarballFile, err := libdownloader.Fetch(sourceURL)
	if err != nil {
		return err
	}

	defer tarballFile.Cleanup()

	logger.Debug.Process("Attempting to verify checksum for version '%s'\n", version)

	actualSha256, err := tarballFile.Sha256()
	if err != nil {
		return err
	}

	if actualSha256 != phpRelease.Sha256 {
		return fmt.Errorf("unable to match checksums for version '%s'.\nExpected '%s'.\nActual '%s'",
			version,
			phpRelease.Sha256,
			actualSha256)
	}

	logger.Debug.Process("Attempting to verify signature for version '%s'\n", version)

	signatureFile, err := libdownloader.Fetch(fmt.Sprintf("%s.asc", sourceURL))
	if err != nil {
		return err
	}

	defer signatureFile.Cleanup()

	signature, err := signatureFile.Contents()
	if err != nil {
		return err
	}

	file, err := os.Open(tarballFile.Path())
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	var pgpKeys []io.Reader
	for constraint, fingerprints := range phpSigningKeys {
		if strings.HasPrefix(version, constraint) {
			for _, fingerprint := range fingerprints {
				signingKeyFile, err := libdownloader.Fetch(getGpgUrl(fingerprint))
				if err != nil {
					return err
				}

				signingKey, err := signingKeyFile.Contents()
				if err != nil {
					return err
				}
				logger.Debug.Subprocess("Found siging key for version '%s' and fingerprint '%s':\n%s\n\n",
					version,
					fingerprint,
					signingKey)
				pgpKeys = append(pgpKeys, strings.NewReader(signingKey))
			}
		}
	}

	verified := false

	var verificationErrors []error

	for _, pgpKey := range pgpKeys {
		keyring, err := openpgp.ReadArmoredKeyRing(pgpKey)
		if err != nil {
			verificationErrors = append(verificationErrors, err)
			continue
		}
		signer, err := openpgp.CheckArmoredDetachedSignature(keyring, file, strings.NewReader(signature))
		if err != nil {
			verificationErrors = append(verificationErrors, err)
			continue
		}
		if signer == nil {
			continue
		}
		verified = true
	}

	if verified {
		logger.Debug.Subprocess("Successfully verified signature for version '%s'\n", version)
		return nil
	}

	for _, err := range verificationErrors {
		logger.Debug.Subprocess("Unable to verify signature for version '%s': %s\n", version, err.Error())
	}

	return fmt.Errorf("unable to verify signature for version '%s'", version)
}
