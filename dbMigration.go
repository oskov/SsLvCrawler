package main

import (
	"database/sql"
	"github.com/GuiaBolso/darwin"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table flats",
			Script: `CREATE TABLE flats (
						id 				INT AUTO_INCREMENT, 
						id_external 	INT,
						text 			VARCHAR(255),
						district    	VARCHAR(100),
						street      	VARCHAR(100),
						rooms 			TINYINT UNSIGNED,
						apartment_area  SMALLINT UNSIGNED,
						floor           VARCHAR(50),
						house_type      VARCHAR(50),
						price			INT,
						PRIMARY KEY (id)
					 ) ENGINE=InnoDB CHARACTER SET=utf8;`,
		},
		{
			Version:     1.1,
			Description: "Add type for sell/rent",
			Script:      `ALTER TABLE flats ADD COLUMN type VARCHAR(4)`,
		},
		{
			Version:     1.2,
			Description: "Unique key by id_external",
			Script:      `CREATE UNIQUE INDEX external_key ON flats(id_external);`,
		},
		{
			Version:     1.3,
			Description: "Add logs",
			Script: `CREATE TABLE logs(
						id 				INT AUTO_INCREMENT, 
						type 			VARCHAR(255),
						log_dt 			DATETIME,
						error 			VARCHAR(255),
						PRIMARY KEY (id)
					) ENGINE=InnoDB CHARACTER SET=utf8;`,
		},
		{
			Version:     1.4,
			Description: "Add url column",
			Script:      `ALTER TABLE flats ADD COLUMN url VARCHAR(255);`,
		},
		{
			Version:     1.5,
			Description: "Add added_dt column",
			Script:      `ALTER TABLE flats ADD COLUMN added_dt DATETIME;`,
		},
		{
			Version:     1.6,
			Description: "Add city column",
			Script:      `ALTER TABLE flats ADD COLUMN city VARCHAR(255) AFTER text;`,
		},
	}
)

func RunMigrations(db *sql.DB) (err error) {
	driver := darwin.NewGenericDriver(db, darwin.MySQLDialect{})
	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()
	return err
}
