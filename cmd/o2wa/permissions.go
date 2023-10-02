package o2wa

func (app *Server) requireGroups(userInfo *authData, requiredGroups []string) bool {
	for _, required := range requiredGroups {
		found := false
		for _, group := range userInfo.IDTokenClaims.Groups {
			if group == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
