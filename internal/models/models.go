package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Player modèle GORM pour la table players
type Player struct {
	ID       string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string    `gorm:"type:text;not null;unique" json:"name"`
	XP       int       `gorm:"not null;default:0" json:"xp"`
	Level    int       `gorm:"not null;default:1" json:"level"`
	Captures []Capture `gorm:"foreignKey:PlayerID" json:"captures,omitempty"`
}

// BeforeCreate génère un UUID avant la création
func (p *Player) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// Word modèle GORM pour la table words
type Word struct {
	ID       string    `gorm:"type:text;primaryKey" json:"id"`
	Text     string    `gorm:"type:text;not null" json:"text"`
	Rarity   string    `gorm:"type:text;not null" json:"rarity"`
	Points   int       `gorm:"not null" json:"points"`
	Captures []Capture `gorm:"foreignKey:WordID" json:"captures,omitempty"`
}

// Capture modèle GORM pour la table captures
type Capture struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	PlayerID   string    `gorm:"type:uuid;not null" json:"player_id"`
	WordID     string    `gorm:"type:text;not null" json:"word_id"`
	CapturedAt time.Time `gorm:"not null;default:now()" json:"captured_at"`

	// Relations
	Player Player `gorm:"foreignKey:PlayerID;constraint:OnDelete:CASCADE" json:"player,omitempty"`
	Word   Word   `gorm:"foreignKey:WordID;constraint:OnDelete:CASCADE" json:"word,omitempty"`
}

// BeforeCreate génère un UUID avant la création
func (c *Capture) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.CapturedAt.IsZero() {
		c.CapturedAt = time.Now()
	}
	return nil
}
