package test

import (
	"testing"
	"io/ioutil"
	"log"	
	"bytes"
	"strings"	
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"	
	"golang.org/x/crypto/ssh"
)

func TestTerraformInstance(t *testing.T) {
	terraformOptions := &terraform.Options{
		// Set the path to the Terraform code that will be tested.
		TerraformDir: "../instance",

		Vars: map[string]interface{}{
			"compartment_ocid": "ocid1.compartment.oc1..aaaaaaaaezxcenaj36sx2s4upu4n76ptsa3mqrkm5ppnu2fswqdpduqywz7q",
			"instance_display_name": "instance-terraform-test",
			"source_ocid": "ocid1.image.oc1.iad.aaaaaaaazvlnvaprv65ak2fqhxzpf6wda2vpbnktiroua3fzrxeizfxzphca",			
			"ssh_authorized_keys": "C:\\Users\\matias.araya.cohen\\OneDrive - Accenture\\Documents\\SSH Keys\\public-key",		
			"subnet_ocids": "[\"ocid1.subnet.oc1.iad.aaaaaaaawnby4ouxffhab45r3qjq2ukbffnmi3lkpi2sabwwvxirl3q7lsca\"]",
			"compute_availability_domain_list" : "[\"MmBO:US-ASHBURN-AD-1\", \"MmBO:US-ASHBURN-AD-2\", \"MmBO:US-ASHBURN-AD-3\"]",
		},

	}

	// Clean up resources with "terraform destroy" at the end of the test.
	defer terraform.Destroy(t, terraformOptions)

	// Run "terraform init" and "terraform apply". Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the values of output variables and check they have the expected values.
	testSSH(t, terraformOptions)	
}


func testSSH(t *testing.T, terraformOptions *terraform.Options){			
	out := terraform.Output(t, terraformOptions, "public_ip")
	ips := strings.Split(out, "\n")
	ips1 := ips[1]
	ips2 := strings.Split(ips1, "\"")
	publicInstanceIP := ips2[1]
	log.Printf(publicInstanceIP)
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile("C:\\Users\\matias.araya.cohen\\OneDrive - Accenture\\Documents\\SSH Keys\\vmInstanceOracleCloud.pem")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}		

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "opc",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		//HostKeyCallback: hostKeyCallback,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}


	maxRetries := 30
	retries:= 0
	timeBetweenRetries := 5 * time.Second	
	
	// Connect to the remote server and perform the SSH handshake.
	for retries < maxRetries{
		retries++
		log.Printf("Trying to connect...")
		client, err := ssh.Dial("tcp", publicInstanceIP+":22", config)
		if err != nil {
			log.Printf("unable to connect: %v", err)
			time.Sleep(timeBetweenRetries)
		} else{
			defer client.Close()
			session, err := client.NewSession()
			if err != nil {
				log.Printf("session failed:%v", err)
			}
			defer session.Close()
			var stdoutBuf bytes.Buffer
			session.Stdout = &stdoutBuf
			err = session.Run("ls -l")
			if err != nil {
				log.Printf("Run failed:%v", err)
			}
			log.Printf(">%s", stdoutBuf.String())
			break
		}		
	}	
}