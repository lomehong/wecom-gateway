// WeCom Gateway Admin UI JavaScript

// API base URL (adjust if needed)
const API_BASE = '/v1';

// JWT Token - set this to authenticate
let jwtToken = localStorage.getItem('wecom_jwt_token') || '';
let currentUser = JSON.parse(localStorage.getItem('wecom_current_user') || 'null');

// Current tab
let currentTab = 'dashboard';

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    // Check if JWT token is set
    if (!jwtToken) {
        showLoginPrompt();
    } else {
        showMainContent();
        loadDashboard();
    }
});

// Show login prompt
function showLoginPrompt() {
    const prompt = document.getElementById('loginPrompt');
    if (prompt) {
        prompt.style.display = 'flex';
    }
    document.getElementById('mainContent').style.display = 'none';
}

// Show main content
function showMainContent() {
    const prompt = document.getElementById('loginPrompt');
    if (prompt) {
        prompt.style.display = 'none';
    }
    document.getElementById('mainContent').style.display = 'block';

    // Update current user display
    if (currentUser) {
        document.getElementById('currentUser').textContent = currentUser.display_name || currentUser.username;
    }
}

// Handle login
async function handleLogin(event) {
    event.preventDefault();

    const username = document.getElementById('usernameInput').value;
    const password = document.getElementById('passwordInput').value;
    const errorDiv = document.getElementById('loginError');

    try {
        const response = await fetch(`${API_BASE}/admin/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        const result = await response.json();

        if (result.code === 0) {
            jwtToken = result.data.access_token;
            localStorage.setItem('wecom_jwt_token', jwtToken);

            currentUser = {
                user_id: result.data.user_id,
                username: result.data.username,
                display_name: result.data.display_name
            };
            localStorage.setItem('wecom_current_user', JSON.stringify(currentUser));

            showMainContent();
            loadDashboard();
        } else {
            errorDiv.textContent = result.message || '登录失败';
            errorDiv.style.display = 'block';
        }
    } catch (error) {
        console.error('Login error:', error);
        errorDiv.textContent = '连接失败，请检查服务是否运行';
        errorDiv.style.display = 'block';
    }
}

// Get auth headers
function getAuthHeaders() {
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${jwtToken}`
    };
}

// Logout function
function logout() {
    if (confirm('确定要退出吗？')) {
        localStorage.removeItem('wecom_jwt_token');
        localStorage.removeItem('wecom_current_user');
        jwtToken = '';
        currentUser = null;
        showLoginPrompt();
    }
}

// Tab switching
function showTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });

    // Remove active class from all buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });

    // Show selected tab
    document.getElementById(tabName).classList.add('active');

    // Set active button
    event.target.classList.add('active');

    currentTab = tabName;

    // Load data for the tab
    switch(tabName) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'corps':
            loadCorps();
            break;
        case 'apps':
            loadApps();
            break;
        case 'keys':
            loadKeys();
            break;
        case 'logs':
            loadLogs();
            break;
    }
}

// Load dashboard data
async function loadDashboard() {
    try {
        const response = await fetch(`${API_BASE}/admin/dashboard`, { headers: getAuthHeaders() });
        const result = await response.json();

        if (result.code === 0) {
            const stats = result.data;
            document.getElementById('totalCorps').textContent = stats.total_corps || 0;
            document.getElementById('totalApps').textContent = stats.total_apps || 0;
            document.getElementById('totalKeys').textContent = stats.total_api_keys || 0;
            document.getElementById('todayRequests').textContent = stats.total_requests || 0;
        } else {
            console.error('Failed to load dashboard:', result.message);
        }
    } catch (error) {
        console.error('Error loading dashboard:', error);
        document.getElementById('totalCorps').textContent = 'Error';
        document.getElementById('totalApps').textContent = 'Error';
        document.getElementById('totalKeys').textContent = 'Error';
        document.getElementById('todayRequests').textContent = 'Error';
    }
}

// Load corps
async function loadCorps() {
    const tbody = document.getElementById('corpsTableBody');
    tbody.innerHTML = '<tr><td colspan="5" class="loading">加载中...</td></tr>';

    try {
        const response = await fetch(`${API_BASE}/admin/corps`, { headers: getAuthHeaders() });
        const result = await response.json();

        if (result.code === 0) {
            const corps = result.data.corps || [];

            if (corps.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: #999;">暂无企业</td></tr>';
                return;
            }

            tbody.innerHTML = corps.map(corp => `
                <tr>
                    <td>${escapeHtml(corp.name)}</td>
                    <td><code>${escapeHtml(corp.corp_id)}</code></td>
                    <td>${corp.app_count || 0}</td>
                    <td>${formatDate(corp.created_at)}</td>
                    <td>
                        <button class="btn btn-small" onclick="showEditCorpModal('${corp.id}', '${escapeHtml(corp.name)}', '${escapeHtml(corp.corp_id)}')">编辑</button>
                        <button class="btn btn-danger btn-small" onclick="deleteCorp('${corp.id}', '${escapeHtml(corp.name)}')">删除</button>
                    </td>
                </tr>
            `).join('');

            // Update corp filter in apps page
            updateCorpFilter(corps);
        } else {
            tbody.innerHTML = `<tr><td colspan="5" class="loading">加载失败: ${escapeHtml(result.message)}</td></tr>`;
        }
    } catch (error) {
        console.error('Error loading corps:', error);
        tbody.innerHTML = '<tr><td colspan="5" class="loading">加载失败: 网络错误</td></tr>';
    }
}

// Update corp filter dropdown
function updateCorpFilter(corps) {
    const corpFilter = document.getElementById('corpFilter');
    const appCorpSelect = document.getElementById('appCorpName');
    const keyCorpSelect = document.getElementById('keyCorp');

    if (corpFilter) {
        corpFilter.innerHTML = '<option value="">所有企业</option>' +
            corps.map(corp => `<option value="${escapeHtml(corp.name)}">${escapeHtml(corp.name)}</option>`).join('');
    }

    if (appCorpSelect) {
        appCorpSelect.innerHTML = '<option value="">请选择企业</option>' +
            corps.map(corp => `<option value="${escapeHtml(corp.name)}">${escapeHtml(corp.name)}</option>`).join('');
    }

    if (keyCorpSelect) {
        keyCorpSelect.innerHTML = '<option value="">请选择企业</option>' +
            corps.map(corp => `<option value="${escapeHtml(corp.name)}">${escapeHtml(corp.name)}</option>`).join('');
    }
}

// Show create corp modal
function showCreateCorpModal() {
    document.getElementById('createCorpModal').classList.add('show');
}

// Close create corp modal
function closeCreateCorpModal() {
    document.getElementById('createCorpModal').classList.remove('show');
    document.getElementById('createCorpForm').reset();
}

// Create corp
async function createCorp(event) {
    event.preventDefault();

    const name = document.getElementById('corpName').value;
    const corpId = document.getElementById('corpId').value;

    try {
        const response = await fetch(`${API_BASE}/admin/corps`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ name, corp_id: corpId })
        });

        const result = await response.json();

        if (result.code === 0) {
            closeCreateCorpModal();
            loadCorps();
            loadDashboard();
        } else {
            alert(`创建失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error creating corp:', error);
        alert('创建失败: 网络错误');
    }
}

// Show edit corp modal
function showEditCorpModal(id, name, corpId) {
    document.getElementById('editCorpId').value = id;
    document.getElementById('editCorpName').value = name;
    document.getElementById('editCorpIdValue').value = corpId;
    document.getElementById('editCorpModal').classList.add('show');
}

// Close edit corp modal
function closeEditCorpModal() {
    document.getElementById('editCorpModal').classList.remove('show');
    document.getElementById('editCorpForm').reset();
}

// Edit corp
async function editCorp(event) {
    event.preventDefault();

    const id = document.getElementById('editCorpId').value;
    const corpId = document.getElementById('editCorpIdValue').value;

    try {
        const response = await fetch(`${API_BASE}/admin/corps/${id}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify({ corp_id: corpId })
        });

        const result = await response.json();

        if (result.code === 0) {
            closeEditCorpModal();
            loadCorps();
        } else {
            alert(`更新失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error updating corp:', error);
        alert('更新失败: 网络错误');
    }
}

// Delete corp
async function deleteCorp(id, name) {
    if (!confirm(`确定要删除企业 "${name}" 吗？此操作将同时删除该企业下的所有应用！`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/corps/${id}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        const result = await response.json();

        if (result.code === 0) {
            loadCorps();
            loadDashboard();
        } else {
            alert(`删除失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error deleting corp:', error);
        alert('删除失败: 网络错误');
    }
}

// Load apps
async function loadApps() {
    const tbody = document.getElementById('appsTableBody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading">加载中...</td></tr>';

    try {
        const corpName = document.getElementById('corpFilter')?.value || '';
        let url = `${API_BASE}/admin/corps/${corpName}/apps`;

        if (!corpName) {
            // Load all corps first, then load apps for each
            const corpsResponse = await fetch(`${API_BASE}/admin/corps`, { headers: getAuthHeaders() });
            const corpsResult = await corpsResponse.json();

            if (corpsResult.code === 0) {
                const corps = corpsResult.data.corps || [];
                const allApps = [];

                for (const corp of corps) {
                    try {
                        const appsResponse = await fetch(`${API_BASE}/admin/corps/${corp.name}/apps`, { headers: getAuthHeaders() });
                        const appsResult = await appsResponse.json();

                        if (appsResult.code === 0) {
                            const apps = appsResult.data.apps || [];
                            allApps.push(...apps);
                        }
                    } catch (error) {
                        console.error(`Error loading apps for corp ${corp.name}:`, error);
                    }
                }

                if (allApps.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="6" style="text-align: center; color: #999;">暂无应用</td></tr>';
                    return;
                }

                tbody.innerHTML = allApps.map(app => `
                    <tr>
                        <td>${escapeHtml(app.name)}</td>
                        <td>${escapeHtml(app.corp_name)}</td>
                        <td><code>${app.agent_id}</code></td>
                        <td>******</td>
                        <td>${formatDate(app.created_at)}</td>
                        <td>
                            <button class="btn btn-small" onclick="showEditAppModal('${app.id}', '${escapeHtml(app.corp_name)}', '${escapeHtml(app.name)}', '${app.agent_id}')">编辑</button>
                            <button class="btn btn-danger btn-small" onclick="deleteApp('${app.id}', '${escapeHtml(app.corp_name)}', '${escapeHtml(app.name)}')">删除</button>
                        </td>
                    </tr>
                `).join('');
            }
        } else {
            const response = await fetch(url, { headers: getAuthHeaders() });
            const result = await response.json();

            if (result.code === 0) {
                const apps = result.data.apps || [];

                if (apps.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="6" style="text-align: center; color: #999;">暂无应用</td></tr>';
                    return;
                }

                tbody.innerHTML = apps.map(app => `
                    <tr>
                        <td>${escapeHtml(app.name)}</td>
                        <td>${escapeHtml(app.corp_name)}</td>
                        <td><code>${app.agent_id}</code></td>
                        <td>******</td>
                        <td>${formatDate(app.created_at)}</td>
                        <td>
                            <button class="btn btn-small" onclick="showEditAppModal('${app.id}', '${escapeHtml(app.corp_name)}', '${escapeHtml(app.name)}', '${app.agent_id}')">编辑</button>
                            <button class="btn btn-danger btn-small" onclick="deleteApp('${app.id}', '${escapeHtml(app.corp_name)}', '${escapeHtml(app.name)}')">删除</button>
                        </td>
                    </tr>
                `).join('');
            }
        }
    } catch (error) {
        console.error('Error loading apps:', error);
        tbody.innerHTML = '<tr><td colspan="6" class="loading">加载失败: 网络错误</td></tr>';
    }
}

// Show create app modal
function showCreateAppModal() {
    document.getElementById('createAppModal').classList.add('show');
}

// Close create app modal
function closeCreateAppModal() {
    document.getElementById('createAppModal').classList.remove('show');
    document.getElementById('createAppForm').reset();
}

// Create app
async function createApp(event) {
    event.preventDefault();

    const corpName = document.getElementById('appCorpName').value;
    const name = document.getElementById('appName').value;
    const agentId = parseInt(document.getElementById('appAgentId').value);
    const secret = document.getElementById('appSecret').value;

    try {
        const response = await fetch(`${API_BASE}/admin/corps/${corpName}/apps`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ name, agent_id: agentId, secret })
        });

        const result = await response.json();

        if (result.code === 0) {
            closeCreateAppModal();
            loadApps();
            loadDashboard();
        } else {
            alert(`创建失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error creating app:', error);
        alert('创建失败: 网络错误');
    }
}

// Show edit app modal
function showEditAppModal(id, corpName, name, agentId) {
    document.getElementById('editAppId').value = id;
    document.getElementById('editAppCorpName').value = corpName;
    document.getElementById('editAppName').value = name;
    document.getElementById('editAppAgentId').value = agentId;
    document.getElementById('editAppSecret').value = '';
    document.getElementById('editAppModal').classList.add('show');
}

// Close edit app modal
function closeEditAppModal() {
    document.getElementById('editAppModal').classList.remove('show');
    document.getElementById('editAppForm').reset();
}

// Edit app
async function editApp(event) {
    event.preventDefault();

    const corpName = document.getElementById('editAppCorpName').value;
    const id = document.getElementById('editAppId').value;
    const name = document.getElementById('editAppName').value;
    const agentId = parseInt(document.getElementById('editAppAgentId').value);
    const secret = document.getElementById('editAppSecret').value;

    const body = { name, agent_id: agentId };
    if (secret) {
        body.secret = secret;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/corps/${corpName}/apps/${id}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify(body)
        });

        const result = await response.json();

        if (result.code === 0) {
            closeEditAppModal();
            loadApps();
        } else {
            alert(`更新失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error updating app:', error);
        alert('更新失败: 网络错误');
    }
}

// Delete app
async function deleteApp(id, corpName, name) {
    if (!confirm(`确定要删除应用 "${name}" 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/corps/${corpName}/apps/${id}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        const result = await response.json();

        if (result.code === 0) {
            loadApps();
            loadDashboard();
        } else {
            alert(`删除失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error deleting app:', error);
        alert('删除失败: 网络错误');
    }
}

// Load API keys
async function loadKeys() {
    const tbody = document.getElementById('keysTableBody');
    tbody.innerHTML = '<tr><td colspan="7" class="loading">加载中...</td></tr>';

    try {
        const response = await fetch(`${API_BASE}/admin/api-keys`, { headers: getAuthHeaders() });
        const result = await response.json();

        if (result.code === 0) {
            const keys = result.data.keys || [];

            if (keys.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #999;">暂无 API Keys</td></tr>';
                return;
            }

            tbody.innerHTML = keys.map(key => `
                <tr>
                    <td>${escapeHtml(key.name)}</td>
                    <td>${key.permissions.map(p => `<code>${escapeHtml(p)}</code>`).join(', ')}</td>
                    <td>${escapeHtml(key.corp_name)}</td>
                    <td>${key.app_name ? escapeHtml(key.app_name) : '-'}</td>
                    <td>
                        <span class="status-badge ${key.disabled ? 'status-disabled' : 'status-active'}">
                            ${key.disabled ? '禁用' : '启用'}
                        </span>
                    </td>
                    <td>${formatDate(key.created_at)}</td>
                    <td>
                        <button class="btn btn-danger btn-small" onclick="deleteKey('${key.id}')">删除</button>
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = `<tr><td colspan="7" class="loading">加载失败: ${escapeHtml(result.message)}</td></tr>`;
        }
    } catch (error) {
        console.error('Error loading keys:', error);
        tbody.innerHTML = '<tr><td colspan="7" class="loading">加载失败: 网络错误</td></tr>';
    }
}

// Show create key modal
function showCreateKeyModal() {
    document.getElementById('createKeyModal').classList.add('show');
}

// Close create key modal
function closeCreateKeyModal() {
    document.getElementById('createKeyModal').classList.remove('show');
    document.getElementById('createKeyForm').reset();
}

// Create API key
async function createKey(event) {
    event.preventDefault();

    const name = document.getElementById('keyName').value;
    const permCheckboxes = document.querySelectorAll('.perm-checkbox:checked');
    const permissions = Array.from(permCheckboxes).map(cb => cb.value);
    const corpName = document.getElementById('keyCorp').value;
    const appName = document.getElementById('keyApp').value;
    const expiresDays = parseInt(document.getElementById('keyExpiry').value) || 0;

    if (permissions.length === 0) {
        alert('请至少选择一个权限');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/api-keys`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                name,
                permissions,
                corp_name: corpName,
                app_name: appName || undefined,
                expires_days: expiresDays
            })
        });

        const result = await response.json();

        if (result.code === 0) {
            document.getElementById('newApiKey').textContent = result.data.api_key;
            document.getElementById('keyCreatedModal').classList.add('show');
            closeCreateKeyModal();
            loadKeys();
            loadDashboard();
        } else {
            alert(`创建失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error creating key:', error);
        alert('创建失败: 网络错误');
    }
}

// Delete key
async function deleteKey(keyId) {
    if (!confirm('确定要删除此 API Key 吗？此操作不可撤销！')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/api-keys/${keyId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        const result = await response.json();

        if (result.code === 0) {
            loadKeys();
            loadDashboard();
        } else {
            alert(`删除失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error deleting key:', error);
        alert('删除失败: 网络错误');
    }
}

// Close key created modal
function closeKeyCreatedModal() {
    document.getElementById('keyCreatedModal').classList.remove('show');
}

// Copy API key to clipboard
function copyApiKey() {
    const keyText = document.getElementById('newApiKey').textContent;

    if (navigator.clipboard) {
        navigator.clipboard.writeText(keyText).then(() => {
            alert('API Key 已复制到剪贴板');
        }).catch(err => {
            console.error('Failed to copy:', err);
            fallbackCopyTextToClipboard(keyText);
        });
    } else {
        fallbackCopyTextToClipboard(keyText);
    }
}

// Fallback copy method
function fallbackCopyTextToClipboard(text) {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.top = "0";
    textArea.style.left = "0";
    textArea.style.width = "2em";
    textArea.style.height = "2em";
    textArea.style.padding = "0";
    textArea.style.border = "none";
    textArea.style.outline = "none";
    textArea.style.boxShadow = "none";
    textArea.style.background = "transparent";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        const successful = document.execCommand('copy');
        if (successful) {
            alert('API Key 已复制到剪贴板');
        } else {
            alert('复制失败，请手动复制');
        }
    } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
        alert('复制失败，请手动复制');
    }

    document.body.removeChild(textArea);
}

// Load audit logs
async function loadLogs() {
    const tbody = document.getElementById('logsTableBody');
    tbody.innerHTML = '<tr><td colspan="7" class="loading">加载中...</td></tr>';

    try {
        const methodFilter = document.getElementById('methodFilter')?.value || '';
        const pathFilter = document.getElementById('pathFilter')?.value || '';

        let url = `${API_BASE}/admin/audit-logs?limit=50`;
        if (methodFilter) url += `&method=${methodFilter}`;
        if (pathFilter) url += `&path=${encodeURIComponent(pathFilter)}`;

        const response = await fetch(url, { headers: getAuthHeaders() });
        const result = await response.json();

        if (result.code === 0) {
            const logs = result.data.logs || [];

            if (logs.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: #999;">暂无日志</td></tr>';
                return;
            }

            tbody.innerHTML = logs.map(log => `
                <tr>
                    <td>${formatDateTime(log.timestamp)}</td>
                    <td>${log.api_key_name ? escapeHtml(log.api_key_name) : '-'}</td>
                    <td><code>${escapeHtml(log.method)}</code></td>
                    <td>${escapeHtml(log.path)}</td>
                    <td>
                        <span style="color: ${log.status_code >= 400 ? '#f5222d' : '#52c41a'}">
                            ${log.status_code}
                        </span>
                    </td>
                    <td>${log.duration_ms}ms</td>
                    <td>${log.client_ip || '-'}</td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = `<tr><td colspan="7" class="loading">加载失败: ${escapeHtml(result.message)}</td></tr>`;
        }
    } catch (error) {
        console.error('Error loading logs:', error);
        tbody.innerHTML = '<tr><td colspan="7" class="loading">加载失败: 网络错误</td></tr>';
    }
}

// Handle change password
async function handleChangePassword(event) {
    event.preventDefault();

    const oldPassword = document.getElementById('oldPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    if (newPassword !== confirmPassword) {
        alert('新密码和确认密码不一致');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/admin/change-password`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                old_password: oldPassword,
                new_password: newPassword
            })
        });

        const result = await response.json();

        if (result.code === 0) {
            alert('密码修改成功');
            document.getElementById('oldPassword').value = '';
            document.getElementById('newPassword').value = '';
            document.getElementById('confirmPassword').value = '';
        } else {
            alert(`密码修改失败: ${result.message}`);
        }
    } catch (error) {
        console.error('Error changing password:', error);
        alert('密码修改失败: 网络错误');
    }
}

// Utility functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleDateString('zh-CN');
}

function formatDateTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN');
}

// Toggle password visibility
function togglePasswordVisibility(inputId, btn) {
    const input = document.getElementById(inputId);
    if (input.type === 'password') {
        input.type = 'text';
        btn.textContent = '🙈';
    } else {
        input.type = 'password';
        btn.textContent = '👁';
    }
}

// Handle "select all" permission checkbox
document.addEventListener('change', function(e) {
    if (e.target.classList.contains('perm-checkbox') && e.target.value === '*') {
        const checkboxes = document.querySelectorAll('.perm-checkbox');
        checkboxes.forEach(cb => {
            cb.checked = e.target.checked;
            if (cb.value !== '*') cb.disabled = e.target.checked;
        });
    } else if (e.target.classList.contains('perm-checkbox') && e.target.value !== '*') {
        const allCheck = document.querySelector('.perm-checkbox[value="*"]');
        if (allCheck && e.target.checked === false) {
            allCheck.checked = false;
            document.querySelectorAll('.perm-checkbox').forEach(cb => cb.disabled = false);
        }
    }
});
