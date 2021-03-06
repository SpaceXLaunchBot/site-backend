package api

import (
	"encoding/hex"
	"github.com/SpaceXLaunchBot/site/internal/encryption"
	"github.com/gin-gonic/gin"
	"time"
)

// removeSessionCookies removes the session cookies for the given context.
func (a Api) removeSessionCookies(c *gin.Context) {
	c.SetCookie("sessionId", "", 0, "/", a.hostName, a.isHTTPS, true)
	c.SetCookie("sessionKey", "", 0, "/", a.hostName, a.isHTTPS, true)
}

// endWithInvalidateSession is similar to endWithResponse but is used to invalidate the users session.
func (a Api) endWithInvalidateSession(c *gin.Context, id string) {
	// TODO: What happens if we remove cookies but fail to remove from db?
	_, _ = a.db.RemoveSession(id)
	a.removeSessionCookies(c)
	endWithResponse(c, responseNoSession)
}

// SessionMiddleware checks for and passes along a decrypted users session as a database.SessionRecord.
func (a Api) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie("sessionId")
		if err != nil {
			endWithResponse(c, responseNoSession)
			return
		}

		sessionKeyHex, err := c.Cookie("sessionKey")
		if err != nil {
			endWithResponse(c, responseNoSession)
			return
		}
		sessionKey, err := hex.DecodeString(sessionKeyHex)
		if err != nil {
			endWithResponse(c, responseInternalError)
			return
		}

		exists, session, err := a.db.GetSession(sessionId)
		if err != nil {
			endWithResponse(c, responseDatabaseError)
			return
		}
		if exists == false {
			endWithResponse(c, responseNoSession)
			return
		}

		// We need to decrypt the stored information, the struct has room for the decrypted as well (below).
		accessTokenBytes, err := encryption.Decrypt(sessionKey, session.AccessTokenEncrypted)
		if err != nil {
			a.endWithInvalidateSession(c, sessionId)
			return
		}

		refreshTokenBytes, err := encryption.Decrypt(sessionKey, session.RefreshTokenEncrypted)
		if err != nil {
			a.endWithInvalidateSession(c, sessionId)
			return
		}

		session.AccessToken = string(accessTokenBytes)
		session.RefreshToken = string(refreshTokenBytes)

		if session.AccessTokenExpiresAt.After(time.Now()) == false {
			// Everything is valid but our access token is expired.
			tokens, err := a.discordClient.RefreshTokens(session.RefreshToken)
			if err != nil {
				a.endWithInvalidateSession(c, sessionId)
				return
			}

			if err = a.encryptAndSetTokens(sessionId, sessionKey, tokens); err != nil {
				a.endWithInvalidateSession(c, sessionId)
				return
			}

			// The get at the top, the set above, and this get, add up to 3 database requests during a single client
			// request. This isn't great but shouldn't happen often.
			exists, session, err = a.db.GetSession(sessionId)
			if err != nil || exists == false {
				endWithResponse(c, responseDatabaseError)
				return
			}
		}

		c.Set("session", session)
		c.Next()
	}
}
