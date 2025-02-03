package internal

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	LINUX_USER        = "edge-extractor"
	LINUX_BIN         = "/usr/local/bin/edge-extractor"
	LINUX_CONFIG_FILE = "/etc/edge-extractor/config.json"
	LINUX_LOG_DIR     = "/var/log/edge-extractor"
)

// function creates edge-extractor linux user and group , copies edge-extractor binary to /usr/local/bin,
// creates config file in /etc/edge-extractor, creates log directory in /var/log/edge-extractor

func PrepareLinuxServiceEnv() error {

	// create edge-extractor user and group
	fmt.Println("1. creating edge-extractor user and group")
	cmd := exec.Command("useradd", "-r", "-s", "/bin/false", LINUX_USER)
	err := cmd.Run()
	if err != nil {
		fmt.Println("1. error creating edge-extractor user , most likely already exists : " + err.Error())
	}
	fmt.Println("1. edge-extractor user and group created")

	// copy edge-extractor binary to /usr/local/bin

	fullBinaryPath, err := os.Executable()
	fmt.Println("2. copying " + fullBinaryPath + " to /usr/local/bin/edge-extractor")
	if err != nil {
		return err
	}
	cmd = exec.Command("cp", "-f", fullBinaryPath, "/usr/local/bin/edge-extractor")
	err = cmd.Run()
	if err != nil {
		fmt.Println("2. error copying edge-extractor binary : " + err.Error())
		return err
	}
	fmt.Println("2. edge-extractor binary copied")

	// create config folder in /etc/edge-extractor
	fmt.Println("3. creating config folder")
	cmd = exec.Command("mkdir", "-p", "/etc/edge-extractor")
	err = cmd.Run()
	if err != nil {
		fmt.Println("3. error creating config folder : " + err.Error())
		return err
	}
	fmt.Println("3. config folder created")

	// check if config file exists
	if _, err := os.Stat("config.json"); !os.IsNotExist(err) {
		// copy config file to /etc/edge-extractor
		fmt.Println("4. config file not found, creating default config file")
		cmd = exec.Command("cp", "config.json", "/etc/edge-extractor")
		err = cmd.Run()
		if err != nil {
			fmt.Println("4. error copying config file : " + err.Error())
			return err
		}
		fmt.Println("4. config file copied")
	} else {
		fmt.Println("4. config file already exists")
	}

	// create log directory in /var/log/edge-extractor
	fmt.Println("5. creating log directory")
	cmd = exec.Command("mkdir", "-p", LINUX_LOG_DIR)
	err = cmd.Run()
	if err != nil {
		fmt.Println("5. error creating log directory : " + err.Error())
		return err
	}
	// change owner of log directory to edge-extractor user
	cmd = exec.Command("chown", "-R", LINUX_USER+":"+LINUX_USER, LINUX_LOG_DIR)
	err = cmd.Run()
	if err != nil {
		fmt.Println("5. error changing owner of log directory : " + err.Error())
		return err
	}
	fmt.Println("5. log directory created")

	return nil
}

// function removes edge-extractor linux user and group , removes edge-extractor binary from /usr/local/bin,
// removes config file from /etc/edge-extractor, removes log directory from /var/log/edge-extractor
func RemoveLinuxServiceEnv() error {

	// remove edge-extractor user and group
	fmt.Println("1. removing edge-extractor user and group")
	cmd := exec.Command("userdel", "-r", LINUX_USER)
	err := cmd.Run()
	if err != nil {
		fmt.Println("1. error removing edge-extractor user : " + err.Error())
	}
	fmt.Println("2. edge-extractor user and group removed")

	// remove edge-extractor binary from /usr/local/bin
	fmt.Println("3. removing edge-extractor binary from /usr/local/bin")
	cmd = exec.Command("rm", "-f", LINUX_BIN)
	err = cmd.Run()
	if err != nil {
		fmt.Println("3. error removing edge-extractor binary : " + err.Error())
	}
	fmt.Println("4. edge-extractor binary removed")

	// remove config file from /etc/edge-extractor
	fmt.Println("5. removing config file from /etc/edge-extractor")
	cmd = exec.Command("rm", "-f", LINUX_CONFIG_FILE)
	err = cmd.Run()
	if err != nil {
		fmt.Println("5. error removing config file : " + err.Error())
	}
	fmt.Println("5. config file removed")

	// remove log directory from /var/log/edge-extractor
	fmt.Println("6. removing log directory from /var/log/edge-extractor")
	cmd = exec.Command("rm", "-rf", LINUX_LOG_DIR)
	err = cmd.Run()
	if err != nil {
		fmt.Println("6. error removing log directory : " + err.Error())
	}
	fmt.Println("6. log directory removed")

	return nil
}

func UpdateLinuxServiceBinary() error {
	fmt.Println("WARNING: This operation will update edge-extractor binary.It might require root privileges or executed using sudo command.")
	// stop edge-extractor service
	fmt.Println("1. stopping edge-extractor service")
	cmd := exec.Command("systemctl", "stop", "edge-extractor")
	err := cmd.Run()
	if err != nil {
		fmt.Println("1. error stopping edge-extractor service . Stop service manually. Error: " + err.Error())
	}
	fmt.Println("2. updating edge-extractor binary")
	fullBinaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	fmt.Println("3. copying " + fullBinaryPath + " to /usr/local/bin/edge-extractor")
	cmd = exec.Command("cp", "-f", fullBinaryPath, "/usr/local/bin/edge-extractor")
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("4. edge-extractor binary updated")
	fmt.Println("5. starting edge-extractor service")
	cmd = exec.Command("systemctl", "start", "edge-extractor")
	err = cmd.Run()
	if err != nil {
		fmt.Println("1. error starting edge-extractor service . Start service manually. Error: " + err.Error())
	}
	return nil
}
