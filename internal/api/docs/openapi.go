package docs

import (
	_ "embed"
)

//go:embed swagger.html
var SwaggerHTML []byte

// OpenAPISpec returns the OpenAPI 3.0 specification as a map
func OpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "GOST Panel API",
			"description": "GOST 代理管理面板 API 文档",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{"url": "/api", "description": "API Server"},
		},
		"components": components(),
		"paths":      paths(),
		"tags":       tags(),
	}
}

func tags() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "Auth", "description": "认证与注册"},
		{"name": "Nodes", "description": "节点管理"},
		{"name": "Clients", "description": "客户端管理"},
		{"name": "Users", "description": "用户管理"},
		{"name": "PortForwards", "description": "端口转发"},
		{"name": "NodeGroups", "description": "节点组/负载均衡"},
		{"name": "ProxyChains", "description": "代理链"},
		{"name": "Tunnels", "description": "隧道转发"},
		{"name": "Rules", "description": "规则管理 (Bypass/Admission/Hosts)"},
		{"name": "Ingresses", "description": "反向代理规则"},
		{"name": "Recorders", "description": "记录器"},
		{"name": "Routers", "description": "路由管理"},
		{"name": "SDs", "description": "服务发现"},
		{"name": "Notify", "description": "通知与告警"},
		{"name": "Plans", "description": "套餐管理"},
		{"name": "Tags", "description": "标签管理"},
		{"name": "Settings", "description": "系统设置"},
		{"name": "Stats", "description": "统计与日志"},
		{"name": "Agent", "description": "Agent 接口"},
	}
}

func ref(name string) map[string]interface{} {
	return map[string]interface{}{"$ref": "#/components/schemas/" + name}
}

func idParam() map[string]interface{} {
	return map[string]interface{}{
		"name": "id", "in": "path", "required": true,
		"schema": map[string]interface{}{"type": "integer"},
		"description": "资源 ID",
	}
}

func jsonBody(schemaRef map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{"schema": schemaRef},
		},
	}
}

func jsonResp(desc string, schemaRef map[string]interface{}) map[string]interface{} {
	resp := map[string]interface{}{"description": desc}
	if schemaRef != nil {
		resp["content"] = map[string]interface{}{
			"application/json": map[string]interface{}{"schema": schemaRef},
		}
	}
	return resp
}

func arrayResp(desc string, schemaRef map[string]interface{}) map[string]interface{} {
	return jsonResp(desc, map[string]interface{}{
		"type":  "array",
		"items": schemaRef,
	})
}

func successResp() map[string]interface{} {
	return jsonResp("成功", ref("SuccessResponse"))
}

func errResp() map[string]interface{} {
	return jsonResp("错误", ref("ErrorResponse"))
}

// crud generates standard CRUD path entries for a resource
func crud(tag, singular, plural, zhName string) map[string]interface{} {
	result := map[string]interface{}{}

	// List + Create
	listCreate := map[string]interface{}{}
	listCreate["get"] = map[string]interface{}{
		"tags": []string{tag}, "summary": "获取" + zhName + "列表",
		"operationId": "list" + plural,
		"responses": map[string]interface{}{
			"200": arrayResp(zhName+"列表", ref(singular)),
		},
	}
	listCreate["post"] = map[string]interface{}{
		"tags": []string{tag}, "summary": "创建" + zhName,
		"operationId":  "create" + singular,
		"requestBody":  jsonBody(ref(singular + "Input")),
		"responses": map[string]interface{}{
			"200": jsonResp("创建成功", ref(singular)),
			"400": errResp(),
		},
	}
	result["/"+lowercase(plural)] = listCreate

	// Get + Update + Delete
	single := map[string]interface{}{}
	single["get"] = map[string]interface{}{
		"tags": []string{tag}, "summary": "获取" + zhName + "详情",
		"operationId": "get" + singular,
		"parameters":  []map[string]interface{}{idParam()},
		"responses": map[string]interface{}{
			"200": jsonResp(zhName+"详情", ref(singular)),
			"404": errResp(),
		},
	}
	single["put"] = map[string]interface{}{
		"tags": []string{tag}, "summary": "更新" + zhName,
		"operationId": "update" + singular,
		"parameters":  []map[string]interface{}{idParam()},
		"requestBody": jsonBody(ref(singular + "Input")),
		"responses": map[string]interface{}{
			"200": successResp(),
			"400": errResp(),
		},
	}
	single["delete"] = map[string]interface{}{
		"tags": []string{tag}, "summary": "删除" + zhName,
		"operationId": "delete" + singular,
		"parameters":  []map[string]interface{}{idParam()},
		"responses": map[string]interface{}{
			"200": successResp(),
		},
	}
	result["/"+lowercase(plural)+"/{id}"] = single

	return result
}

func lowercase(s string) string {
	if len(s) == 0 {
		return s
	}
	result := make([]byte, 0, len(s)+4)
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, byte(c+'a'-'A'))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

func merge(maps ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			if existing, ok := result[k]; ok {
				if existingMap, ok2 := existing.(map[string]interface{}); ok2 {
					if newMap, ok3 := v.(map[string]interface{}); ok3 {
						for mk, mv := range newMap {
							existingMap[mk] = mv
						}
						continue
					}
				}
			}
			result[k] = v
		}
	}
	return result
}

func paths() map[string]interface{} {
	p := merge(
		authPaths(),
		crud("Nodes", "Node", "Nodes", "节点"),
		nodePaths(),
		crud("Clients", "Client", "Clients", "客户端"),
		clientPaths(),
		crud("Users", "User", "Users", "用户"),
		userPaths(),
		crud("PortForwards", "PortForward", "PortForwards", "端口转发"),
		crud("NodeGroups", "NodeGroup", "NodeGroups", "节点组"),
		nodeGroupPaths(),
		crud("ProxyChains", "ProxyChain", "ProxyChains", "代理链"),
		proxyChainPaths(),
		crud("Tunnels", "Tunnel", "Tunnels", "隧道"),
		tunnelPaths(),
		crud("Rules", "Bypass", "Bypasses", "分流规则"),
		crud("Rules", "Admission", "Admissions", "准入规则"),
		hostMappingPaths(),
		crud("Ingresses", "Ingress", "Ingresses", "反向代理规则"),
		crud("Recorders", "Recorder", "Recorders", "记录器"),
		crud("Routers", "Router", "Routers", "路由"),
		crud("SDs", "SD", "SDs", "服务发现"),
		crud("Notify", "NotifyChannel", "NotifyChannels", "通知渠道"),
		notifyPaths(),
		crud("Plans", "Plan", "Plans", "套餐"),
		crud("Tags", "Tag", "Tags", "标签"),
		tagPaths(),
		statsPaths(),
		settingsPaths(),
		agentPaths(),
	)
	return p
}

func authPaths() map[string]interface{} {
	return map[string]interface{}{
		"/login": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Auth"}, "summary": "用户登录",
				"operationId": "login",
				"requestBody": jsonBody(map[string]interface{}{
					"type":     "object",
					"required": []string{"username", "password"},
					"properties": map[string]interface{}{
						"username": map[string]interface{}{"type": "string"},
						"password": map[string]interface{}{"type": "string"},
					},
				}),
				"responses": map[string]interface{}{
					"200": jsonResp("登录成功", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"token": map[string]interface{}{"type": "string"},
							"user":  ref("User"),
						},
					}),
					"401": errResp(),
				},
			},
		},
		"/register": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Auth"}, "summary": "用户注册",
				"operationId": "register",
				"requestBody": jsonBody(map[string]interface{}{
					"type":     "object",
					"required": []string{"username", "password", "email"},
					"properties": map[string]interface{}{
						"username": map[string]interface{}{"type": "string"},
						"password": map[string]interface{}{"type": "string"},
						"email":    map[string]interface{}{"type": "string", "format": "email"},
					},
				}),
				"responses": map[string]interface{}{
					"200": successResp(),
					"400": errResp(),
				},
			},
		},
		"/change-password": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Auth"}, "summary": "修改密码",
				"operationId": "changePassword",
				"security":    []map[string]interface{}{{"bearerAuth": []string{}}},
				"requestBody": jsonBody(map[string]interface{}{
					"type":     "object",
					"required": []string{"old_password", "new_password"},
					"properties": map[string]interface{}{
						"old_password": map[string]interface{}{"type": "string"},
						"new_password": map[string]interface{}{"type": "string"},
					},
				}),
				"responses": map[string]interface{}{
					"200": successResp(),
					"400": errResp(),
				},
			},
		},
		"/profile": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Auth"}, "summary": "获取个人信息",
				"operationId": "getProfile",
				"responses": map[string]interface{}{
					"200": jsonResp("个人信息", ref("User")),
				},
			},
			"put": map[string]interface{}{
				"tags": []string{"Auth"}, "summary": "更新个人信息",
				"operationId": "updateProfile",
				"requestBody": jsonBody(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"email": map[string]interface{}{"type": "string", "format": "email"},
					},
				}),
				"responses": map[string]interface{}{
					"200": successResp(),
				},
			},
		},
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "健康检查",
				"operationId": "healthCheck",
				"security":    []map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": jsonResp("服务健康", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"status": map[string]interface{}{"type": "string", "example": "ok"},
						},
					}),
				},
			},
		},
	}
}

func nodePaths() map[string]interface{} {
	return map[string]interface{}{
		"/nodes/paginated": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "分页获取节点列表",
				"operationId": "listNodesPaginated",
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 1}},
					{"name": "page_size", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
				},
				"responses": map[string]interface{}{
					"200": jsonResp("分页节点列表", ref("PaginatedResponse")),
				},
			},
		},
		"/nodes/{id}/apply": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "应用节点配置到 GOST",
				"operationId": "applyNodeConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/nodes/{id}/sync": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "同步节点配置",
				"operationId": "syncNodeConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/nodes/{id}/gost-config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "获取节点 GOST 配置",
				"operationId": "getNodeGostConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "GOST YAML 配置"},
				},
			},
		},
		"/nodes/{id}/proxy-uri": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "获取节点代理 URI",
				"operationId": "getNodeProxyURI",
				"parameters":  []map[string]interface{}{idParam()},
				"responses": map[string]interface{}{
					"200": jsonResp("代理 URI", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"uri": map[string]interface{}{"type": "string"},
						},
					}),
				},
			},
		},
		"/nodes/{id}/install-script": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "获取节点安装脚本",
				"operationId": "getNodeInstallScript",
				"parameters":  []map[string]interface{}{idParam()},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "安装脚本"},
				},
			},
		},
		"/nodes/{id}/ping": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "Ping 单个节点",
				"operationId": "pingNode",
				"parameters":  []map[string]interface{}{idParam()},
				"responses": map[string]interface{}{
					"200": jsonResp("延迟结果", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"latency_ms": map[string]interface{}{"type": "number"},
						},
					}),
				},
			},
		},
		"/nodes/ping": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "Ping 所有节点",
				"operationId": "pingAllNodes",
				"responses": map[string]interface{}{
					"200": jsonResp("所有节点延迟结果", map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{"type": "number"},
					}),
				},
			},
		},
		"/nodes/batch-delete": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "批量删除节点",
				"operationId": "batchDeleteNodes",
				"requestBody": jsonBody(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}},
					},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/nodes/batch-sync": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Nodes"}, "summary": "批量同步节点",
				"operationId": "batchSyncNodes",
				"requestBody": jsonBody(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}},
					},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/nodes/{id}/tags": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Tags"}, "summary": "获取节点标签",
				"operationId": "getNodeTags",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": arrayResp("标签列表", ref("Tag"))},
			},
			"post": map[string]interface{}{
				"tags": []string{"Tags"}, "summary": "添加节点标签",
				"operationId": "addNodeTag",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"tag_id": map[string]interface{}{"type": "integer"}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
			"put": map[string]interface{}{
				"tags": []string{"Tags"}, "summary": "设置节点标签 (全量替换)",
				"operationId": "setNodeTags",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"tag_ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/nodes/{id}/tags/{tagId}": map[string]interface{}{
			"delete": map[string]interface{}{
				"tags": []string{"Tags"}, "summary": "移除节点标签",
				"operationId": "removeNodeTag",
				"parameters": []map[string]interface{}{
					idParam(),
					{"name": "tagId", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
	}
}

func clientPaths() map[string]interface{} {
	return map[string]interface{}{
		"/clients/paginated": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Clients"}, "summary": "分页获取客户端列表",
				"operationId": "listClientsPaginated",
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 1}},
					{"name": "page_size", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
				},
				"responses": map[string]interface{}{"200": jsonResp("分页列表", ref("PaginatedResponse"))},
			},
		},
		"/clients/batch-delete": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Clients"}, "summary": "批量删除客户端",
				"operationId": "batchDeleteClients",
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/clients/{id}/install-script": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Clients"}, "summary": "获取客户端安装脚本",
				"operationId": "getClientInstallScript",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "安装脚本"}},
			},
		},
		"/clients/{id}/gost-config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Clients"}, "summary": "获取客户端 GOST 配置",
				"operationId": "getClientGostConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "GOST YAML 配置"}},
			},
		},
		"/clients/{id}/proxy-uri": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Clients"}, "summary": "获取客户端代理 URI",
				"operationId": "getClientProxyURI",
				"parameters":  []map[string]interface{}{idParam()},
				"responses": map[string]interface{}{
					"200": jsonResp("代理 URI", map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"uri": map[string]interface{}{"type": "string"}},
					}),
				},
			},
		},
	}
}

func userPaths() map[string]interface{} {
	return map[string]interface{}{
		"/users/{id}/verify-email": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Users"}, "summary": "管理员验证用户邮箱",
				"operationId": "adminVerifyUserEmail",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/users/{id}/reset-quota": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Users"}, "summary": "重置用户流量配额",
				"operationId": "resetUserQuota",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/users/{id}/assign-plan": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Users"}, "summary": "分配套餐给用户",
				"operationId": "assignUserPlan",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"plan_id": map[string]interface{}{"type": "integer"}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/users/{id}/remove-plan": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Users"}, "summary": "移除用户套餐",
				"operationId": "removeUserPlan",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/users/{id}/renew-plan": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Users"}, "summary": "续期用户套餐",
				"operationId": "renewUserPlan",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"days": map[string]interface{}{"type": "integer"}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
	}
}

func nodeGroupPaths() map[string]interface{} {
	return map[string]interface{}{
		"/node-groups/{id}/members": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"NodeGroups"}, "summary": "获取节点组成员",
				"operationId": "listNodeGroupMembers",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": arrayResp("成员列表", ref("NodeGroupMember"))},
			},
			"post": map[string]interface{}{
				"tags": []string{"NodeGroups"}, "summary": "添加节点组成员",
				"operationId": "addNodeGroupMember",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"node_id": map[string]interface{}{"type": "integer"}, "weight": map[string]interface{}{"type": "integer"}},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/node-groups/{id}/members/{memberId}": map[string]interface{}{
			"delete": map[string]interface{}{
				"tags": []string{"NodeGroups"}, "summary": "移除节点组成员",
				"operationId": "removeNodeGroupMember",
				"parameters": []map[string]interface{}{
					idParam(),
					{"name": "memberId", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/node-groups/{id}/config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"NodeGroups"}, "summary": "获取节点组配置",
				"operationId": "getNodeGroupConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "节点组配置"}},
			},
		},
	}
}

func proxyChainPaths() map[string]interface{} {
	return map[string]interface{}{
		"/proxy-chains/{id}/hops": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"ProxyChains"}, "summary": "获取代理链跳板列表",
				"operationId": "listProxyChainHops",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": arrayResp("跳板列表", ref("ProxyChainHop"))},
			},
			"post": map[string]interface{}{
				"tags": []string{"ProxyChains"}, "summary": "添加代理链跳板",
				"operationId": "addProxyChainHop",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(ref("ProxyChainHopInput")),
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/proxy-chains/{id}/hops/{hopId}": map[string]interface{}{
			"put": map[string]interface{}{
				"tags": []string{"ProxyChains"}, "summary": "更新代理链跳板",
				"operationId": "updateProxyChainHop",
				"parameters": []map[string]interface{}{
					idParam(),
					{"name": "hopId", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"requestBody": jsonBody(ref("ProxyChainHopInput")),
				"responses":   map[string]interface{}{"200": successResp()},
			},
			"delete": map[string]interface{}{
				"tags": []string{"ProxyChains"}, "summary": "删除代理链跳板",
				"operationId": "removeProxyChainHop",
				"parameters": []map[string]interface{}{
					idParam(),
					{"name": "hopId", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/proxy-chains/{id}/config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"ProxyChains"}, "summary": "获取代理链配置",
				"operationId": "getProxyChainConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "代理链配置"}},
			},
		},
	}
}

func tunnelPaths() map[string]interface{} {
	return map[string]interface{}{
		"/tunnels/{id}/sync": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Tunnels"}, "summary": "同步隧道配置",
				"operationId": "syncTunnel",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/tunnels/{id}/entry-config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Tunnels"}, "summary": "获取隧道入口配置",
				"operationId": "getTunnelEntryConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "入口 GOST 配置"}},
			},
		},
		"/tunnels/{id}/exit-config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Tunnels"}, "summary": "获取隧道出口配置",
				"operationId": "getTunnelExitConfig",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "出口 GOST 配置"}},
			},
		},
	}
}

func hostMappingPaths() map[string]interface{} {
	p := crud("Rules", "HostMapping", "HostMappings", "主机映射")
	// Fix the path key to use hyphenated form
	fixed := map[string]interface{}{}
	for k, v := range p {
		switch k {
		case "/host-mappings":
			fixed["/host-mappings"] = v
		case "/host-mappings/{id}":
			fixed["/host-mappings/{id}"] = v
		default:
			fixed[k] = v
		}
	}
	return fixed
}

func notifyPaths() map[string]interface{} {
	return map[string]interface{}{
		"/notify-channels/{id}/test": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "测试通知渠道",
				"operationId": "testNotifyChannel",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/alert-rules": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "获取告警规则列表",
				"operationId": "listAlertRules",
				"responses":   map[string]interface{}{"200": arrayResp("告警规则列表", ref("AlertRule"))},
			},
			"post": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "创建告警规则",
				"operationId": "createAlertRule",
				"requestBody": jsonBody(ref("AlertRuleInput")),
				"responses":   map[string]interface{}{"200": jsonResp("创建成功", ref("AlertRule"))},
			},
		},
		"/alert-rules/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "获取告警规则详情",
				"operationId": "getAlertRule",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": jsonResp("告警规则", ref("AlertRule"))},
			},
			"put": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "更新告警规则",
				"operationId": "updateAlertRule",
				"parameters":  []map[string]interface{}{idParam()},
				"requestBody": jsonBody(ref("AlertRuleInput")),
				"responses":   map[string]interface{}{"200": successResp()},
			},
			"delete": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "删除告警规则",
				"operationId": "deleteAlertRule",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": successResp()},
			},
		},
		"/alert-logs": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Notify"}, "summary": "获取告警日志",
				"operationId": "getAlertLogs",
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
					{"name": "page_size", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{"200": jsonResp("告警日志", ref("PaginatedResponse"))},
			},
		},
	}
}

func tagPaths() map[string]interface{} {
	return map[string]interface{}{
		"/tags/{id}/nodes": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Tags"}, "summary": "获取标签下的节点",
				"operationId": "getNodesByTag",
				"parameters":  []map[string]interface{}{idParam()},
				"responses":   map[string]interface{}{"200": arrayResp("节点列表", ref("Node"))},
			},
		},
	}
}

func statsPaths() map[string]interface{} {
	return map[string]interface{}{
		"/stats": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Stats"}, "summary": "获取面板统计数据",
				"operationId": "getStats",
				"responses": map[string]interface{}{
					"200": jsonResp("统计数据", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"total_nodes":    map[string]interface{}{"type": "integer"},
							"online_nodes":   map[string]interface{}{"type": "integer"},
							"total_clients":  map[string]interface{}{"type": "integer"},
							"online_clients": map[string]interface{}{"type": "integer"},
							"total_users":    map[string]interface{}{"type": "integer"},
							"total_traffic":  map[string]interface{}{"type": "integer"},
						},
					}),
				},
			},
		},
		"/search": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Stats"}, "summary": "全局搜索",
				"operationId": "globalSearch",
				"parameters": []map[string]interface{}{
					{"name": "q", "in": "query", "required": true, "schema": map[string]interface{}{"type": "string"}, "description": "搜索关键词"},
				},
				"responses": map[string]interface{}{"200": jsonResp("搜索结果", map[string]interface{}{"type": "object"})},
			},
		},
		"/traffic-history": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Stats"}, "summary": "获取流量历史",
				"operationId": "getTrafficHistory",
				"parameters": []map[string]interface{}{
					{"name": "node_id", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
					{"name": "hours", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 24}},
				},
				"responses": map[string]interface{}{"200": arrayResp("流量历史", ref("TrafficHistory"))},
			},
		},
		"/operation-logs": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Stats"}, "summary": "获取操作日志",
				"operationId": "getOperationLogs",
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
					{"name": "page_size", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
					{"name": "action", "in": "query", "schema": map[string]interface{}{"type": "string"}},
				},
				"responses": map[string]interface{}{"200": jsonResp("操作日志", ref("PaginatedResponse"))},
			},
		},
		"/export": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "导出数据",
				"operationId": "exportData",
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "JSON 数据文件"}},
			},
		},
		"/import": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "导入数据",
				"operationId": "importData",
				"requestBody": map[string]interface{}{
					"content": map[string]interface{}{
						"multipart/form-data": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":       "object",
								"properties": map[string]interface{}{"file": map[string]interface{}{"type": "string", "format": "binary"}},
							},
						},
					},
				},
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/backup": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "备份数据库",
				"operationId": "backupDatabase",
				"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "SQLite 数据库文件"}},
			},
		},
		"/restore": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "恢复数据库",
				"operationId": "restoreDatabase",
				"requestBody": map[string]interface{}{
					"content": map[string]interface{}{
						"multipart/form-data": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":       "object",
								"properties": map[string]interface{}{"file": map[string]interface{}{"type": "string", "format": "binary"}},
							},
						},
					},
				},
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
	}
}

func settingsPaths() map[string]interface{} {
	return map[string]interface{}{
		"/site-configs": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "获取站点配置",
				"operationId": "getSiteConfigs",
				"responses":   map[string]interface{}{"200": jsonResp("站点配置列表", map[string]interface{}{"type": "array", "items": ref("SiteConfig")})},
			},
			"put": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "更新站点配置",
				"operationId": "updateSiteConfigs",
				"requestBody": jsonBody(map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"key": map[string]interface{}{"type": "string"}, "value": map[string]interface{}{"type": "string"}},
					},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/site-config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Settings"}, "summary": "获取公开站点配置",
				"operationId": "getPublicSiteConfig",
				"security":    []map[string]interface{}{},
				"responses":   map[string]interface{}{"200": jsonResp("公开配置", map[string]interface{}{"type": "object"})},
			},
		},
	}
}

func agentPaths() map[string]interface{} {
	return map[string]interface{}{
		"/agent/register": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Agent"}, "summary": "Agent 注册",
				"operationId": "agentRegister",
				"security":    []map[string]interface{}{},
				"requestBody": jsonBody(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token":   map[string]interface{}{"type": "string"},
						"version": map[string]interface{}{"type": "string"},
						"os":      map[string]interface{}{"type": "string"},
						"arch":    map[string]interface{}{"type": "string"},
					},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/agent/heartbeat": map[string]interface{}{
			"post": map[string]interface{}{
				"tags": []string{"Agent"}, "summary": "Agent 心跳",
				"operationId": "agentHeartbeat",
				"security":    []map[string]interface{}{},
				"requestBody": jsonBody(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token":          map[string]interface{}{"type": "string"},
						"upload_bytes":   map[string]interface{}{"type": "integer"},
						"download_bytes": map[string]interface{}{"type": "integer"},
					},
				}),
				"responses": map[string]interface{}{"200": successResp()},
			},
		},
		"/agent/config/{token}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Agent"}, "summary": "获取 Agent 配置",
				"operationId": "agentGetConfig",
				"security":    []map[string]interface{}{},
				"parameters": []map[string]interface{}{
					{"name": "token", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
				},
				"responses": map[string]interface{}{"200": map[string]interface{}{"description": "GOST YAML 配置"}},
			},
		},
		"/agent/version": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Agent"}, "summary": "获取 Agent 最新版本",
				"operationId": "agentGetVersion",
				"security":    []map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": jsonResp("版本信息", map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"version": map[string]interface{}{"type": "string"}, "build_time": map[string]interface{}{"type": "string"}},
					}),
				},
			},
		},
		"/agent/check-update": map[string]interface{}{
			"get": map[string]interface{}{
				"tags": []string{"Agent"}, "summary": "检查 Agent 更新",
				"operationId": "agentCheckUpdate",
				"security":    []map[string]interface{}{},
				"parameters": []map[string]interface{}{
					{"name": "version", "in": "query", "schema": map[string]interface{}{"type": "string"}},
				},
				"responses": map[string]interface{}{
					"200": jsonResp("更新信息", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"has_update":      map[string]interface{}{"type": "boolean"},
							"latest_version":  map[string]interface{}{"type": "string"},
							"current_version": map[string]interface{}{"type": "string"},
						},
					}),
				},
			},
		},
	}
}

func components() map[string]interface{} {
	return map[string]interface{}{
		"securitySchemes": map[string]interface{}{
			"bearerAuth": map[string]interface{}{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
			},
		},
		"schemas": schemas(),
	}
}

func schemas() map[string]interface{} {
	return map[string]interface{}{
		"SuccessResponse": map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"success": map[string]interface{}{"type": "boolean"}},
		},
		"ErrorResponse": map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"error": map[string]interface{}{"type": "string"}},
		},
		"PaginatedResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "object"}},
				"total":      map[string]interface{}{"type": "integer"},
				"page":       map[string]interface{}{"type": "integer"},
				"page_size":  map[string]interface{}{"type": "integer"},
				"total_pages": map[string]interface{}{"type": "integer"},
			},
		},
		"Node": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "integer"},
				"name":      map[string]interface{}{"type": "string"},
				"host":      map[string]interface{}{"type": "string"},
				"port":      map[string]interface{}{"type": "integer"},
				"protocol":  map[string]interface{}{"type": "string", "enum": []string{"socks5", "socks4", "http", "http2", "ss", "ssu", "auto", "relay", "tcp", "udp", "sni", "dns", "sshd", "redirect", "redu", "tun", "tap"}},
				"transport": map[string]interface{}{"type": "string"},
				"username":  map[string]interface{}{"type": "string"},
				"password":  map[string]interface{}{"type": "string"},
				"is_online": map[string]interface{}{"type": "boolean"},
				"upload_bytes":   map[string]interface{}{"type": "integer"},
				"download_bytes": map[string]interface{}{"type": "integer"},
				"token":     map[string]interface{}{"type": "string"},
				"owner_id":  map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"NodeInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "host", "port"},
			"properties": map[string]interface{}{
				"name":      map[string]interface{}{"type": "string"},
				"host":      map[string]interface{}{"type": "string"},
				"port":      map[string]interface{}{"type": "integer"},
				"protocol":  map[string]interface{}{"type": "string", "enum": []string{"socks5", "socks4", "http", "http2", "ss", "ssu", "auto", "relay", "tcp", "udp", "sni", "dns", "sshd", "redirect", "redu", "tun", "tap"}},
				"transport": map[string]interface{}{"type": "string"},
				"username":  map[string]interface{}{"type": "string"},
				"password":  map[string]interface{}{"type": "string"},
			},
		},
		"Client": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "integer"},
				"name":      map[string]interface{}{"type": "string"},
				"node_id":   map[string]interface{}{"type": "integer"},
				"listen_port": map[string]interface{}{"type": "integer"},
				"is_online": map[string]interface{}{"type": "boolean"},
				"upload_bytes":   map[string]interface{}{"type": "integer"},
				"download_bytes": map[string]interface{}{"type": "integer"},
				"token":     map[string]interface{}{"type": "string"},
				"owner_id":  map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"ClientInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "node_id"},
			"properties": map[string]interface{}{
				"name":      map[string]interface{}{"type": "string"},
				"node_id":   map[string]interface{}{"type": "integer"},
				"listen_port": map[string]interface{}{"type": "integer"},
			},
		},
		"User": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "integer"},
				"username":  map[string]interface{}{"type": "string"},
				"email":     map[string]interface{}{"type": "string"},
				"is_admin":  map[string]interface{}{"type": "boolean"},
				"traffic_quota": map[string]interface{}{"type": "integer"},
				"traffic_used":  map[string]interface{}{"type": "integer"},
				"plan_id":   map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"UserInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"username", "password"},
			"properties": map[string]interface{}{
				"username": map[string]interface{}{"type": "string"},
				"password": map[string]interface{}{"type": "string"},
				"email":    map[string]interface{}{"type": "string"},
				"is_admin": map[string]interface{}{"type": "boolean"},
				"traffic_quota": map[string]interface{}{"type": "integer"},
			},
		},
		"PortForward": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":          map[string]interface{}{"type": "integer"},
				"name":        map[string]interface{}{"type": "string"},
				"node_id":     map[string]interface{}{"type": "integer"},
				"protocol":    map[string]interface{}{"type": "string", "enum": []string{"tcp", "udp", "rtcp", "rudp"}},
				"listen_port": map[string]interface{}{"type": "integer"},
				"target_addr": map[string]interface{}{"type": "string"},
				"target_port": map[string]interface{}{"type": "integer"},
				"owner_id":    map[string]interface{}{"type": "integer"},
				"created_at":  map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"PortForwardInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "node_id", "protocol", "listen_port", "target_addr", "target_port"},
			"properties": map[string]interface{}{
				"name":        map[string]interface{}{"type": "string"},
				"node_id":     map[string]interface{}{"type": "integer"},
				"protocol":    map[string]interface{}{"type": "string", "enum": []string{"tcp", "udp", "rtcp", "rudp"}},
				"listen_port": map[string]interface{}{"type": "integer"},
				"target_addr": map[string]interface{}{"type": "string"},
				"target_port": map[string]interface{}{"type": "integer"},
			},
		},
		"NodeGroup": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"name":     map[string]interface{}{"type": "string"},
				"strategy": map[string]interface{}{"type": "string", "enum": []string{"round", "random", "fifo", "hash"}},
				"owner_id": map[string]interface{}{"type": "integer"},
			},
		},
		"NodeGroupInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":     map[string]interface{}{"type": "string"},
				"strategy": map[string]interface{}{"type": "string", "enum": []string{"round", "random", "fifo", "hash"}},
			},
		},
		"NodeGroupMember": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"group_id": map[string]interface{}{"type": "integer"},
				"node_id":  map[string]interface{}{"type": "integer"},
				"weight":   map[string]interface{}{"type": "integer"},
			},
		},
		"ProxyChain": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"name":     map[string]interface{}{"type": "string"},
				"owner_id": map[string]interface{}{"type": "integer"},
			},
		},
		"ProxyChainInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
			},
		},
		"ProxyChainHop": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"chain_id": map[string]interface{}{"type": "integer"},
				"node_id":  map[string]interface{}{"type": "integer"},
				"order":    map[string]interface{}{"type": "integer"},
			},
		},
		"ProxyChainHopInput": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"node_id": map[string]interface{}{"type": "integer"},
				"order":   map[string]interface{}{"type": "integer"},
			},
		},
		"Tunnel": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":            map[string]interface{}{"type": "integer"},
				"name":          map[string]interface{}{"type": "string"},
				"entry_node_id": map[string]interface{}{"type": "integer"},
				"exit_node_id":  map[string]interface{}{"type": "integer"},
				"listen_port":   map[string]interface{}{"type": "integer"},
				"target_addr":   map[string]interface{}{"type": "string"},
				"target_port":   map[string]interface{}{"type": "integer"},
				"owner_id":      map[string]interface{}{"type": "integer"},
			},
		},
		"TunnelInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "entry_node_id", "exit_node_id"},
			"properties": map[string]interface{}{
				"name":          map[string]interface{}{"type": "string"},
				"entry_node_id": map[string]interface{}{"type": "integer"},
				"exit_node_id":  map[string]interface{}{"type": "integer"},
				"listen_port":   map[string]interface{}{"type": "integer"},
				"target_addr":   map[string]interface{}{"type": "string"},
				"target_port":   map[string]interface{}{"type": "integer"},
			},
		},
		"Bypass": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "integer"},
				"name":      map[string]interface{}{"type": "string"},
				"whitelist": map[string]interface{}{"type": "boolean"},
				"matchers":  map[string]interface{}{"type": "string", "description": "JSON 数组: [\"*.google.com\", \"10.0.0.0/8\"]"},
				"node_id":   map[string]interface{}{"type": "integer"},
				"owner_id":  map[string]interface{}{"type": "integer"},
			},
		},
		"BypassInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":      map[string]interface{}{"type": "string"},
				"whitelist": map[string]interface{}{"type": "boolean"},
				"matchers":  map[string]interface{}{"type": "string"},
				"node_id":   map[string]interface{}{"type": "integer"},
			},
		},
		"Admission": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "integer"},
				"name":      map[string]interface{}{"type": "string"},
				"whitelist": map[string]interface{}{"type": "boolean"},
				"matchers":  map[string]interface{}{"type": "string", "description": "JSON 数组: [\"192.168.0.0/16\"]"},
				"node_id":   map[string]interface{}{"type": "integer"},
				"owner_id":  map[string]interface{}{"type": "integer"},
			},
		},
		"AdmissionInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":      map[string]interface{}{"type": "string"},
				"whitelist": map[string]interface{}{"type": "boolean"},
				"matchers":  map[string]interface{}{"type": "string"},
				"node_id":   map[string]interface{}{"type": "integer"},
			},
		},
		"HostMapping": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"name":     map[string]interface{}{"type": "string"},
				"mappings": map[string]interface{}{"type": "string", "description": "JSON 数组: [{\"hostname\":\"example.com\",\"ip\":\"1.2.3.4\"}]"},
				"node_id":  map[string]interface{}{"type": "integer"},
				"owner_id": map[string]interface{}{"type": "integer"},
			},
		},
		"HostMappingInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":     map[string]interface{}{"type": "string"},
				"mappings": map[string]interface{}{"type": "string"},
				"node_id":  map[string]interface{}{"type": "integer"},
			},
		},
		"Ingress": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":         map[string]interface{}{"type": "integer"},
				"name":       map[string]interface{}{"type": "string"},
				"rules":      map[string]interface{}{"type": "string", "description": "JSON 数组: [{\"hostname\":\"example.com\",\"endpoint\":\"192.168.1.1:8080\"}]"},
				"node_id":    map[string]interface{}{"type": "integer"},
				"owner_id":   map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"IngressInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"rules":   map[string]interface{}{"type": "string"},
				"node_id": map[string]interface{}{"type": "integer"},
			},
		},
		"Recorder": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":         map[string]interface{}{"type": "integer"},
				"name":       map[string]interface{}{"type": "string"},
				"type":       map[string]interface{}{"type": "string", "enum": []string{"file", "redis", "http"}},
				"config":     map[string]interface{}{"type": "string", "description": "JSON 配置"},
				"node_id":    map[string]interface{}{"type": "integer"},
				"owner_id":   map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"RecorderInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "type"},
			"properties": map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"type":    map[string]interface{}{"type": "string", "enum": []string{"file", "redis", "http"}},
				"config":  map[string]interface{}{"type": "string"},
				"node_id": map[string]interface{}{"type": "integer"},
			},
		},
		"Router": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":         map[string]interface{}{"type": "integer"},
				"name":       map[string]interface{}{"type": "string"},
				"routes":     map[string]interface{}{"type": "string", "description": "JSON 数组: [{\"net\":\"192.168.1.0/24\",\"gateway\":\"10.0.0.1\"}]"},
				"node_id":    map[string]interface{}{"type": "integer"},
				"owner_id":   map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"RouterInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"routes":  map[string]interface{}{"type": "string"},
				"node_id": map[string]interface{}{"type": "integer"},
			},
		},
		"SD": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":         map[string]interface{}{"type": "integer"},
				"name":       map[string]interface{}{"type": "string"},
				"type":       map[string]interface{}{"type": "string", "enum": []string{"http", "consul", "etcd", "redis"}},
				"config":     map[string]interface{}{"type": "string", "description": "JSON 配置"},
				"node_id":    map[string]interface{}{"type": "integer"},
				"owner_id":   map[string]interface{}{"type": "integer"},
				"created_at": map[string]interface{}{"type": "string", "format": "date-time"},
				"updated_at": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"SDInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "type"},
			"properties": map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"type":    map[string]interface{}{"type": "string", "enum": []string{"http", "consul", "etcd", "redis"}},
				"config":  map[string]interface{}{"type": "string"},
				"node_id": map[string]interface{}{"type": "integer"},
			},
		},
		"NotifyChannel": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "integer"},
				"name":     map[string]interface{}{"type": "string"},
				"type":     map[string]interface{}{"type": "string", "enum": []string{"telegram", "webhook", "smtp"}},
				"config":   map[string]interface{}{"type": "string", "description": "JSON 配置"},
				"enabled":  map[string]interface{}{"type": "boolean"},
			},
		},
		"NotifyChannelInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "type", "config"},
			"properties": map[string]interface{}{
				"name":    map[string]interface{}{"type": "string"},
				"type":    map[string]interface{}{"type": "string", "enum": []string{"telegram", "webhook", "smtp"}},
				"config":  map[string]interface{}{"type": "string"},
				"enabled": map[string]interface{}{"type": "boolean"},
			},
		},
		"AlertRule": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":         map[string]interface{}{"type": "integer"},
				"name":       map[string]interface{}{"type": "string"},
				"type":       map[string]interface{}{"type": "string"},
				"threshold":  map[string]interface{}{"type": "number"},
				"channel_ids": map[string]interface{}{"type": "string"},
				"enabled":    map[string]interface{}{"type": "boolean"},
			},
		},
		"AlertRuleInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name", "type"},
			"properties": map[string]interface{}{
				"name":       map[string]interface{}{"type": "string"},
				"type":       map[string]interface{}{"type": "string"},
				"threshold":  map[string]interface{}{"type": "number"},
				"channel_ids": map[string]interface{}{"type": "string"},
				"enabled":    map[string]interface{}{"type": "boolean"},
			},
		},
		"Plan": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":            map[string]interface{}{"type": "integer"},
				"name":          map[string]interface{}{"type": "string"},
				"traffic_limit": map[string]interface{}{"type": "integer"},
				"node_limit":    map[string]interface{}{"type": "integer"},
				"duration_days": map[string]interface{}{"type": "integer"},
				"price":         map[string]interface{}{"type": "number"},
			},
		},
		"PlanInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":          map[string]interface{}{"type": "string"},
				"traffic_limit": map[string]interface{}{"type": "integer"},
				"node_limit":    map[string]interface{}{"type": "integer"},
				"duration_days": map[string]interface{}{"type": "integer"},
				"price":         map[string]interface{}{"type": "number"},
			},
		},
		"Tag": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":    map[string]interface{}{"type": "integer"},
				"name":  map[string]interface{}{"type": "string"},
				"color": map[string]interface{}{"type": "string"},
			},
		},
		"TagInput": map[string]interface{}{
			"type":     "object",
			"required": []string{"name"},
			"properties": map[string]interface{}{
				"name":  map[string]interface{}{"type": "string"},
				"color": map[string]interface{}{"type": "string"},
			},
		},
		"TrafficHistory": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":             map[string]interface{}{"type": "integer"},
				"node_id":        map[string]interface{}{"type": "integer"},
				"upload_bytes":   map[string]interface{}{"type": "integer"},
				"download_bytes": map[string]interface{}{"type": "integer"},
				"recorded_at":    map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"SiteConfig": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":    map[string]interface{}{"type": "integer"},
				"key":   map[string]interface{}{"type": "string"},
				"value": map[string]interface{}{"type": "string"},
			},
		},
	}
}
