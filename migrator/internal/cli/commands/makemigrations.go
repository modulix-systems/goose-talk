package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func MakeMigrations(migrationsPath string) {
	var migrationName, addMigrationBodyReply string
	var err error
	fmt.Print("Enter migration name: ")
	if _, err = fmt.Scanln(&migrationName); err != nil {
		panic(err)
	}
	fmt.Print("Do you want to add migrations body? (yes/no)[no] ")
	if addMigrationBodyReply, err = bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		panic(err)
	}
	migrationVersion := time.Now().Unix()
	migrationUp := createMigrationFile(migrationName, int(migrationVersion), "up", migrationsPath)
	migrationDown := createMigrationFile(migrationName, int(migrationVersion), "down", migrationsPath)
	if strings.ToLower(strings.TrimSpace(addMigrationBodyReply)) == "yes" {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			fmt.Println("WARNING: editor not set, using vi")
			editor = "vi"
		}
		launchEditor(editor, migrationUp.Name())
		launchEditor(editor, migrationDown.Name())
	}
	fmt.Println("Migrations succesfully created!")
}

func launchEditor(executableName string, filename string) {
	cmd := exec.Command(executableName, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func createMigrationFile(title string, versionNum int, direction string, migrationsPath string) *os.File {
	filename := fmt.Sprintf("%d_%s.%s.sql", versionNum, title, direction)
	file, err := os.Create(path.Join(migrationsPath, filename))
	if err != nil {
		panic(err)
	}
	return file
}
