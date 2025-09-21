package main

import (
	"fmt"
	"os"
	
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/args"
)

func main() {
	// Setup EVO first so configuration is loaded
	evo.Setup()
	
	// Initialize models after setup
	InitializeModels()
	
	// Check for test mode using EVO args system
	if args.Exists("--test") {
		fmt.Println("🧪 Running in Test Mode")
		TestMigration()
		fmt.Println("🔬 Running Comprehensive Tests")
		TestSpecialCases()
		fmt.Println("🔄 Testing Foreign Key Formats")
		TestForeignKeyFormats()
		fmt.Println("🔄 Testing Schema Changes")
		TestSchemaChanges()
		os.Exit(0)
	}
	
	fmt.Println("🌟 Starting EVO Development Migration Server")
	fmt.Println("Database models loaded and ready for migration testing")
	
	// Run the server
	evo.Run()
}
