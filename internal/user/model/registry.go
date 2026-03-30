package model

import "gorm.io/gorm"

func getAllModels() []interface{} {
	return []interface{}{
		&User{},
		&Relation{},
		// &NewTable{}, // Future models go here
	}
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(getAllModels()...); err != nil {
		return err
	}
	return nil
}
