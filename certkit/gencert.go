package certkit

import (
   "os"
   "fmt"
   "time"
   "errors"
   "strings"
   "math/big"
   "crypto/rsa"
   "crypto/rand"
   "crypto/x509"
   "encoding/pem"
   "encoding/asn1"
   "crypto/x509/pkix"
)



func New(srvsubject, casubject pkix.Name, host, email string) (*CertKit, error) {
   var crtkit CertKit
   var e      error

   crtkit = CertKit{}

   e = crtkit.GenerateCA(casubject, host, email)
   if  e != nil {
      return nil, e
   }

   e = crtkit.GenerateServer(srvsubject, host, email)
   if  e != nil {
      return nil, e
   }

   return &crtkit, nil
}


func (crtkit *CertKit) GenerateServer(subject pkix.Name, host, email string) error {
   var e          error
   var derBytes []byte

   priv, err := rsa.GenerateKey(rand.Reader, 2048)
   if err != nil {
      return errors.New(fmt.Sprintf("failed to generate private key: %s", err))
   }

   notBefore         := time.Now()
   serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
   if err != nil {
      return errors.New(fmt.Sprintf("failed to generate serial number: %s", err))
   }

   if host == "" {
      host, _ = os.Hostname()
   }

   Goose.Logf(4,"Certificate authority used: %#v", crtkit.CACert)

   template := x509.Certificate{
      SerialNumber:          serialNumber,
      Subject:               subject,
      IsCA:                  false,
      NotBefore:             notBefore,
      NotAfter:              notBefore.Add(365*24*time.Hour),
      DNSNames:              []string{host, strings.Split(host,".")[0]},
      AuthorityKeyId:        crtkit.CACert.SubjectKeyId,
      KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment,
      ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
      BasicConstraintsValid: true,
   }

   if crtkit.CACert.CRLDistributionPoints != nil {
      template.CRLDistributionPoints = crtkit.CACert.CRLDistributionPoints
   } else {
      Goose.Logf(1,"Certificate authority without CRL distribution points")
   }

   crtkit.ServerKey        = priv
   crtkit.ServerCert       = &template
   derBytes, e             = x509.CreateCertificate(rand.Reader, &template, crtkit.CACert, &priv.PublicKey, crtkit.CAKey)
   if e != nil {
      return errors.New(fmt.Sprintf("Failed to create certificate: %s", e))
   }
   crtkit.ServerCertPem    = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
   crtkit.ServerKeyPem     = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

   return nil
}

func (crtkit *CertKit) GenerateCA(subject pkix.Name, host, email string, listenport ...string) error {
   var e          error
   var derBytes []byte
   var crlurl     string
//   var ecBytes  []byte

//   priv, err := ecdsa.GenerateKey(elliptic.P521(),rand.Reader)
   priv, err := rsa.GenerateKey(rand.Reader, 2048)
   if err != nil {
      return errors.New(fmt.Sprintf("failed to generate private key: %s", err))
   }

   notBefore         := time.Now()
   serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
   if err != nil {
      return errors.New(fmt.Sprintf("failed to generate serial number: %s", err))
   }

   if host == "" {
      host, _ = os.Hostname()
   }

   if len(listenport) == 0 {
      crlurl = "http://" + host + "/rootCA.crl"
   } else {
      crlurl = "http://" + host + ":" + listenport[0] + "/rootCA.crl"
   }

   template := x509.Certificate{
      SerialNumber:          serialNumber,
      Subject:               subject,
      IsCA:                  true,
      SubjectKeyId:          []byte(fmt.Sprintf("%s",priv.PublicKey.N)),
      AuthorityKeyId:        []byte(fmt.Sprintf("%s",priv.PublicKey.N)),

      NotBefore:             notBefore,
      NotAfter:              notBefore.Add(365*20*24*time.Hour),
      DNSNames:              []string{host, strings.Split(host,".")[0]},
      PolicyIdentifiers:     []asn1.ObjectIdentifier{[]int{2, 16, 76, 1, 1, 0}}, // Policy: 2.16.76.1.1.0 CPS: http://acraiz.icpbrasil.gov.br/DPCacraiz.pdf
      CRLDistributionPoints: []string{crlurl},

      KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
      BasicConstraintsValid: true,
   }

   crtkit.CAKey                = priv
   crtkit.CACert               = &template
   derBytes, e                 = x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
   if e != nil {
      return errors.New(fmt.Sprintf("Failed to create certificate: %s", e))
   }
   crtkit.CACertPem            = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
//   ecBytes, e                  = x509.MarshalECPrivateKey(priv)
//   if err != nil {
//      return errors.New(fmt.Sprintf("Failed to Marshal ECDSA Private Key: %s", e))
//   }
//   crtkit.CAKeyPem             = pem.EncodeToMemory(&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: ecBytes})
   crtkit.CAKeyPem             = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

   derBytes, e = template.CreateCRL(rand.Reader, priv, []pkix.RevokedCertificate{}, time.Now(), time.Now().Add(time.Hour*24*30))
   if e != nil {
      return errors.New(fmt.Sprintf("Failed to create CRL: %s", e))
   }
   crtkit.CACRL                = derBytes

   return nil
}




func (crtkit *CertKit) GenerateClient(subject pkix.Name, email, password string) ([]byte,[]byte,error) {
   priv, err := rsa.GenerateKey(rand.Reader, 2048)
   if err != nil {
      return nil, nil, errors.New(fmt.Sprintf("failed to generate private key: %s", err))
   }

   notBefore         := time.Now()
   serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
   if err != nil {
      return nil, nil, errors.New(fmt.Sprintf("failed to generate serial number: %s", err))
   }

   template := x509.Certificate{
      SerialNumber:          serialNumber,
      Subject:               subject,
      NotBefore:             notBefore,
      NotAfter:              notBefore.Add(3650*24*time.Hour),
      EmailAddresses:        []string{email},
      KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
      ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
      BasicConstraintsValid: true,
   }


   derBytes, err := x509.CreateCertificate(rand.Reader, &template, crtkit.CACert, &priv.PublicKey, crtkit.CAKey)
   if err != nil {
      return nil, nil, errors.New(fmt.Sprintf("Failed to create certificate: %s", err))
   }

   certOut := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

   crypt_priv, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(priv), []byte(password), x509.PEMCipher3DES)
   if err != nil {
      return nil, nil, errors.New(fmt.Sprintf("Failed to encrypt: %s", err))
   }

   keyOut  := pem.EncodeToMemory(crypt_priv)

   return certOut, keyOut, nil
}
