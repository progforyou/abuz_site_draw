package data

import (
	"bot_tasker/shared/axcrudobject"
	"errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

type Reward struct {
	axcrudobject.Model
	Ip    string
	Hash  string
	Timer time.Time
	Can   bool `gorm:"-"`
}

func (i *Reward) AfterFind(tx *gorm.DB) (err error) {
	i.Can = !i.Timer.Before(time.Now())
	return
}

const TIME = 24

type RewardController struct {
	Set    func(string, string) error
	Get    func(string, string) (Reward, error)
	Create func(string, string) error
	Check  func(string, string) bool
}

func NewRewardController(db *gorm.DB, baseLog zerolog.Logger) RewardController {
	log := baseLog.With().Str("model", "reward").Logger()
	if err := db.AutoMigrate(&Reward{}); err != nil {
		log.Fatal().Err(err).Msg("auto-migrate")
	}
	return RewardController{
		Set: func(ip, session string) error {
			var obj Reward
			var oldObj Reward
			obj.Ip = ip
			obj.Hash = session
			obj.Timer = time.Now().Add(time.Hour * TIME)
			if err := db.Where("ip = ? OR hash = ?", ip, session).Find(&oldObj).Error; err != nil {
				return err
			}
			if oldObj.Timer.Before(time.Now()) {
				if tx := db.Where("ip = ? AND hash = ?", ip, session).Find(&oldObj); tx.RowsAffected == 0 {
					tx := db.Save(&obj)
					if tx.Error != nil {
						log.Error().Err(tx.Error).Msg("db update error")
						return tx.Error
					}
				} else {
					obj.ID = oldObj.ID
					tx := db.Save(&obj)
					if tx.Error != nil {
						log.Error().Err(tx.Error).Msg("db update error")
						return tx.Error
					}
				}
			} else {
				return errors.New("wait for time")
			}
			return nil
		},
		Check: func(ip, session string) bool {
			var obj Reward
			if err := db.Where("ip = ? OR hash = ?", ip, session).Find(&obj).Error; err != nil {
				return false
			}
			if obj.ID > 0 {
				return true
			}
			return false
		},
		Create: func(ip, session string) error {
			var obj Reward
			obj.Ip = ip
			obj.Hash = session
			obj.Timer = time.Now().Add(time.Hour * TIME)
			tx := db.Model(&Reward{}).Create(&obj)
			if tx.Error != nil {
				log.Error().Err(tx.Error).Msg("db error")
				return tx.Error
			}
			return nil
		},
		Get: func(ip string, session string) (Reward, error) {
			var obj Reward
			if err := db.Where("ip = ? AND hash = ?", ip, session).Find(&obj).Error; err != nil {
				return Reward{}, err
			}
			return obj, nil
		},
	}
}
