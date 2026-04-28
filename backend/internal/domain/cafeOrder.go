package domain

type CafeOrder struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	EnergyCost    int64  `json:"energy_cost"`
	RewardXP      int64  `json:"reward_xp"`
	RequiredLevel int64  `json:"required_level"`
}
