package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Benefit represents a benefit with multiple redemption codes
type Benefit struct {
	ID               uint        `json:"id" gorm:"primaryKey"`
	UUID             string      `json:"uuid" gorm:"type:varchar(255);uniqueIndex"` // For generating private links
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	CreatorID        uint        `json:"creator_id"`
	Creator          User        `json:"-" gorm:"foreignKey:CreatorID"`
	TotalCount       int         `json:"total_count"`
	ClaimedCount     int         `json:"claimed_count"`
	CreatedAt        time.Time   `json:"created_at"`
	ExpiresAt        time.Time   `json:"expires_at"`
	Status           string      `json:"status" gorm:"default:'active'"` // active/paused/expired/deleted
	AllowedProviders StringSlice `json:"allowed_providers" gorm:"type:json"`
	MinAccountAge    int         `json:"min_account_age"`
	ClaimConditions  JSON        `json:"claim_conditions" gorm:"type:json"`
}

// RedemptionCode represents a single code within a benefit
type RedemptionCode struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	BenefitID uint       `json:"benefit_id"`
	Benefit   Benefit    `json:"-" gorm:"foreignKey:BenefitID"`
	Code      string     `json:"code"`
	Status    string     `json:"status" gorm:"default:'available'"` // available/claimed/expired
	ClaimedBy *uint      `json:"claimed_by"`                        // 使用指针类型，允许为NULL
	User      User       `json:"-" gorm:"foreignKey:ClaimedBy"`
	ClaimedAt *time.Time `json:"claimed_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// Claim represents a record of a user claiming a benefit
type Claim struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"index:idx_user_benefit,unique:true"`
	User           User           `json:"-" gorm:"foreignKey:UserID"`
	BenefitID      uint           `json:"benefit_id" gorm:"index:idx_user_benefit,unique:true"`
	Benefit        Benefit        `json:"-" gorm:"foreignKey:BenefitID"`
	CodeID         uint           `json:"code_id"`
	RedemptionCode RedemptionCode `json:"-" gorm:"foreignKey:CodeID"`
	OAuthProvider  string         `json:"oauth_provider"`
	ClaimedAt      time.Time      `json:"claimed_at"` // 这个字段始终有值，不需要是指针
	IPAddress      string         `json:"ip_address"`
	UserAgent      string         `json:"user_agent"`
}

// StringSlice is a custom type for string slices in the database
type StringSlice []string

// Scan implements the sql.Scanner interface
func (ss *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ss)
}

// Value implements the driver.Valuer interface
func (ss StringSlice) Value() (driver.Value, error) {
	return json.Marshal(ss)
}

// JSON is a custom type for storing JSON data
type JSON map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}
