package pogifyapi

type SessionClaim struct {
	SessionID string `json:"sessionId" binding:"required"`
	Issued    Time   `json:"issued" binding:"required"`
	Checksum  string `json:"checksum" binding:"required"`
	Solution  string `json:"solution" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
}
