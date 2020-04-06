package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

// GET /api/user  (current authenticated user)
func GetSignedInUser(c *m.ReqContext) Response {
	return getUserUserProfile(c.UserId)
}

// GET /api/users/:id
func GetUserByID(c *m.ReqContext) Response {
	return getUserUserProfile(c.ParamsInt64(":id"))
}

func getUserUserProfile(userID int64) Response {
	query := m.GetUserProfileQuery{UserId: userID}

	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrUserNotFound {
			return Error(404, m.ErrUserNotFound.Error(), nil)
		}
		return Error(500, "无法获得用户", err)
	}

	getAuthQuery := m.GetAuthInfoQuery{UserId: userID}
	query.Result.AuthLabels = []string{}
	if err := bus.Dispatch(&getAuthQuery); err == nil {
		authLabel := GetAuthProviderLabel(getAuthQuery.Result.AuthModule)
		query.Result.AuthLabels = append(query.Result.AuthLabels, authLabel)
		query.Result.IsExternal = true
	}

	return JSON(200, query.Result)
}

// GET /api/users/lookup
func GetUserByLoginOrEmail(c *m.ReqContext) Response {
	query := m.GetUserByLoginQuery{LoginOrEmail: c.Query("loginOrEmail")}
	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrUserNotFound {
			return Error(404, m.ErrUserNotFound.Error(), nil)
		}
		return Error(500, "无法获得用户", err)
	}
	user := query.Result
	result := m.UserProfileDTO{
		Id:             user.Id,
		Name:           user.Name,
		Email:          user.Email,
		Login:          user.Login,
		Theme:          user.Theme,
		IsGrafanaAdmin: user.IsAdmin,
		OrgId:          user.OrgId,
	}
	return JSON(200, &result)
}

// POST /api/user
func UpdateSignedInUser(c *m.ReqContext, cmd m.UpdateUserCommand) Response {
	if setting.AuthProxyEnabled {
		if setting.AuthProxyHeaderProperty == "email" && cmd.Email != c.Email {
			return Error(400, "当验证代理使用电子邮件属性时，不允许更改电子邮件", nil)
		}
		if setting.AuthProxyHeaderProperty == "username" && cmd.Login != c.Login {
			return Error(400, "当验证代理使用用户名属性时，不允许更改用户名", nil)
		}
	}
	cmd.UserId = c.UserId
	return handleUpdateUser(cmd)
}

// POST /api/users/:id
func UpdateUser(c *m.ReqContext, cmd m.UpdateUserCommand) Response {
	cmd.UserId = c.ParamsInt64(":id")
	return handleUpdateUser(cmd)
}

//POST /api/users/:id/using/:orgId
func UpdateUserActiveOrg(c *m.ReqContext) Response {
	userID := c.ParamsInt64(":id")
	orgID := c.ParamsInt64(":orgId")

	if !validateUsingOrg(userID, orgID) {
		return Error(401, "不是一个有效的组织", nil)
	}

	cmd := m.SetUsingOrgCommand{UserId: userID, OrgId: orgID}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更改活动的组织", err)
	}

	return Success("组织改变了")
}

func handleUpdateUser(cmd m.UpdateUserCommand) Response {
	if len(cmd.Login) == 0 {
		cmd.Login = cmd.Email
		if len(cmd.Login) == 0 {
			return Error(400, "验证错误，需要指定用户名或电子邮件", nil)
		}
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更新用户", err)
	}

	return Success("用户更新完成")
}

// GET /api/user/orgs
func GetSignedInUserOrgList(c *m.ReqContext) Response {
	return getUserOrgList(c.UserId)
}

// GET /api/user/teams
func GetSignedInUserTeamList(c *m.ReqContext) Response {
	return getUserTeamList(c.OrgId, c.UserId)
}

// GET /api/users/:id/teams
func GetUserTeams(c *m.ReqContext) Response {
	return getUserTeamList(c.OrgId, c.ParamsInt64(":id"))
}

func getUserTeamList(orgID int64, userID int64) Response {
	query := m.GetTeamsByUserQuery{OrgId: orgID, UserId: userID}

	if err := bus.Dispatch(&query); err != nil {
		return Error(500, "无法获得用户团队", err)
	}

	for _, team := range query.Result {
		team.AvatarUrl = dtos.GetGravatarUrlWithDefault(team.Email, team.Name)
	}
	return JSON(200, query.Result)
}

// GET /api/users/:id/orgs
func GetUserOrgList(c *m.ReqContext) Response {
	return getUserOrgList(c.ParamsInt64(":id"))
}

func getUserOrgList(userID int64) Response {
	query := m.GetUserOrgListQuery{UserId: userID}

	if err := bus.Dispatch(&query); err != nil {
		return Error(500, "无法获得用户团队", err)
	}

	return JSON(200, query.Result)
}

func validateUsingOrg(userID int64, orgID int64) bool {
	query := m.GetUserOrgListQuery{UserId: userID}

	if err := bus.Dispatch(&query); err != nil {
		return false
	}

	// validate that the org id in the list
	valid := false
	for _, other := range query.Result {
		if other.OrgId == orgID {
			valid = true
		}
	}

	return valid
}

// POST /api/user/using/:id
func UserSetUsingOrg(c *m.ReqContext) Response {
	orgID := c.ParamsInt64(":id")

	if !validateUsingOrg(c.UserId, orgID) {
		return Error(401, "不是一个有效的团队", nil)
	}

	cmd := m.SetUsingOrgCommand{UserId: c.UserId, OrgId: orgID}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更改活动团队", err)
	}

	return Success("积极团队改变了")
}

// GET /profile/switch-org/:id
func (hs *HTTPServer) ChangeActiveOrgAndRedirectToHome(c *m.ReqContext) {
	orgID := c.ParamsInt64(":id")

	if !validateUsingOrg(c.UserId, orgID) {
		hs.NotFoundHandler(c)
	}

	cmd := m.SetUsingOrgCommand{UserId: c.UserId, OrgId: orgID}

	if err := bus.Dispatch(&cmd); err != nil {
		hs.NotFoundHandler(c)
	}

	c.Redirect(setting.AppSubUrl + "/")
}

func ChangeUserPassword(c *m.ReqContext, cmd m.ChangeUserPasswordCommand) Response {
	if setting.LDAPEnabled || setting.AuthProxyEnabled {
		return Error(400, "启用LDAP或验证代理时，不允许更改密码", nil)
	}

	userQuery := m.GetUserByIdQuery{Id: c.UserId}

	if err := bus.Dispatch(&userQuery); err != nil {
		return Error(500, "无法从数据库中读取用户", err)
	}

	passwordHashed := util.EncodePassword(cmd.OldPassword, userQuery.Result.Salt)
	if passwordHashed != userQuery.Result.Password {
		return Error(401, "旧密码无效", nil)
	}

	password := m.Password(cmd.NewPassword)
	if password.IsWeak() {
		return Error(400, "新密码太短", nil)
	}

	cmd.UserId = c.UserId
	cmd.NewPassword = util.EncodePassword(cmd.NewPassword, userQuery.Result.Salt)

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更改用户密码", err)
	}

	return Success("用户密码已变更")
}

// GET /api/users
func SearchUsers(c *m.ReqContext) Response {
	query, err := searchUser(c)
	if err != nil {
		return Error(500, "无法获取用户", err)
	}

	return JSON(200, query.Result.Users)
}

// GET /api/users/search
func SearchUsersWithPaging(c *m.ReqContext) Response {
	query, err := searchUser(c)
	if err != nil {
		return Error(500, "无法获取用户", err)
	}

	return JSON(200, query.Result)
}

func searchUser(c *m.ReqContext) (*m.SearchUsersQuery, error) {
	perPage := c.QueryInt("perpage")
	if perPage <= 0 {
		perPage = 1000
	}
	page := c.QueryInt("page")

	if page < 1 {
		page = 1
	}

	searchQuery := c.Query("query")

	query := &m.SearchUsersQuery{Query: searchQuery, Page: page, Limit: perPage}
	if err := bus.Dispatch(query); err != nil {
		return nil, err
	}

	for _, user := range query.Result.Users {
		user.AvatarUrl = dtos.GetGravatarUrl(user.Email)
		user.AuthLabels = make([]string, 0)
		if user.AuthModule != nil && len(user.AuthModule) > 0 {
			for _, authModule := range user.AuthModule {
				user.AuthLabels = append(user.AuthLabels, GetAuthProviderLabel(authModule))
			}
		}
	}

	query.Result.Page = page
	query.Result.PerPage = perPage

	return query, nil
}

func SetHelpFlag(c *m.ReqContext) Response {
	flag := c.ParamsInt64(":id")

	bitmask := &c.HelpFlags1
	bitmask.AddFlag(m.HelpFlags1(flag))

	cmd := m.SetUserHelpFlagCommand{
		UserId:     c.UserId,
		HelpFlags1: *bitmask,
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更新帮助标志", err)
	}

	return JSON(200, &util.DynMap{"message": "帮助标志设置", "helpFlags1": cmd.HelpFlags1})
}

func ClearHelpFlags(c *m.ReqContext) Response {
	cmd := m.SetUserHelpFlagCommand{
		UserId:     c.UserId,
		HelpFlags1: m.HelpFlags1(0),
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更新帮助标志", err)
	}

	return JSON(200, &util.DynMap{"message": "帮助标志设置", "helpFlags1": cmd.HelpFlags1})
}

func GetAuthProviderLabel(authModule string) string {
	switch authModule {
	case "oauth_github":
		return "GitHub"
	case "oauth_google":
		return "Google"
	case "oauth_gitlab":
		return "GitLab"
	case "oauth_grafana_com", "oauth_grafananet":
		return "grafana.com"
	case "auth.saml":
		return "SAML"
	case "ldap", "":
		return "LDAP"
	default:
		return "OAuth"
	}
}
