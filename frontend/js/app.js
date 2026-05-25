// ===== Shared App Logic =====

// ===== Auth =====

function initAuthState() {
  const user = getCurrentUser();
  const btnLogin = document.getElementById('btnLogin');
  const btnRegister = document.getElementById('btnRegister');
  const userMenu = document.getElementById('userMenu');
  const nameDisplay = document.getElementById('userNameDisplay');

  if (!btnLogin && !userMenu) return;

  if (user) {
    if (btnLogin) btnLogin.style.display = 'none';
    if (btnRegister) btnRegister.style.display = 'none';
    if (userMenu) {
      userMenu.style.display = 'flex';
      if (nameDisplay) nameDisplay.textContent = user.username;
    }
  } else {
    if (userMenu) userMenu.style.display = 'none';
    if (btnLogin) btnLogin.style.display = '';
    if (btnRegister) btnRegister.style.display = '';
  }
}

function logout() {
  localStorage.removeItem('user');
  showToast('已退出登录', 'success');
  setTimeout(() => { window.location.reload(); }, 500);
}

// ===== Toast =====

function showToast(message, type = 'success') {
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);
  setTimeout(() => { toast.remove(); }, 3000);
}

// ===== Activity List =====

function renderActivities(filter) {
  const grid = document.getElementById('activityGrid');
  if (!grid) return;

  let activities = mockGetActivities();
  if (filter !== 'all') {
    activities = activities.filter(a => a.status === filter);
  }

  if (activities.length === 0) {
    grid.innerHTML = '<div class="empty-state"><div class="empty-icon">📭</div><h3>暂无活动</h3><p>请稍后再来看看</p></div>';
    return;
  }

  const statusMap = {
    active: { label: '进行中', cls: 'badge-success' },
    drawed: { label: '已开奖', cls: 'badge-info' },
    canceled: { label: '已取消', cls: 'badge-danger' },
  };

  grid.innerHTML = activities.map(a => {
    const s = statusMap[a.status];
    const actionBtn = a.status === 'active'
      ? '<a href="activity.html?id=' + a.id + '" class="btn btn-primary btn-sm">立即报名</a>'
      : a.status === 'drawed'
        ? '<a href="activity.html?id=' + a.id + '" class="btn btn-secondary btn-sm">查看结果</a>'
        : '<button class="btn btn-secondary btn-sm" disabled>已取消</button>';

    return `
      <div class="card activity-card" onclick="location.href='activity.html?id=${a.id}'">
        <div class="card-header">
          <div class="card-title">${a.name}</div>
          <span class="badge ${s.cls}"><span class="status-dot ${a.status}"></span>${s.label}</span>
        </div>
        <div class="card-meta">
          <span>📅 ${a.startTime} ~ ${a.endTime}</span>
        </div>
        <div class="card-stats">
          <div class="stat">
            <div class="stat-value">${a.participantCount.toLocaleString()}</div>
            <div class="stat-label">参与人数</div>
          </div>
          <div class="stat">
            <div class="stat-value">${a.luckyCount}</div>
            <div class="stat-label">中奖名额</div>
          </div>
        </div>
        <div class="card-footer">
          <span class="text-sm text-secondary">合约名: ${a.contractName}</span>
          ${actionBtn}
        </div>
      </div>
    `;
  }).join('');
}

// ===== Activity Detail =====

function renderActivityDetail(id) {
  const activity = mockGetActivity(id);
  if (!activity) {
    document.getElementById('activityTitle').textContent = '活动不存在';
    return;
  }

  document.getElementById('activityTitle').textContent = activity.name;
  document.getElementById('breadcrumbName').textContent = activity.name;

  const statusMap = {
    active: { label: '进行中', cls: 'badge-success' },
    drawed: { label: '已开奖', cls: 'badge-info' },
    canceled: { label: '已取消', cls: 'badge-danger' },
  };
  const s = statusMap[activity.status];

  document.getElementById('detailInfo').innerHTML = `
    <div class="detail-info-item">
      <div class="label">状态</div>
      <div class="value"><span class="badge ${s.cls}">${s.label}</span></div>
    </div>
    <div class="detail-info-item">
      <div class="label">开始时间</div>
      <div class="value" style="font-size:1rem;">${activity.startTime}</div>
    </div>
    <div class="detail-info-item">
      <div class="label">截止时间</div>
      <div class="value" style="font-size:1rem;">${activity.endTime}</div>
    </div>
    <div class="detail-info-item">
      <div class="label">参与人数</div>
      <div class="value">${activity.participantCount.toLocaleString()}</div>
    </div>
    <div class="detail-info-item">
      <div class="label">中奖名额</div>
      <div class="value text-gold">${activity.luckyCount}</div>
    </div>
  `;

  // Register button for active activities
  if (activity.status === 'active') {
    document.getElementById('registerSection').style.display = 'block';
  }

  // Participant table
  const participantTable = document.getElementById('participantTable');
  participantTable.innerHTML = activity.participants.map((p, i) => `
    <tr>
      <td><span class="font-mono">#${p.index}</span></td>
      <td>${p.username}</td>
      <td><span class="font-mono">${p.address}</span></td>
      <td class="text-sm text-secondary">${p.time}</td>
    </tr>
  `).join('');
  document.getElementById('participantCount').textContent =
    `显示前 ${activity.participants.length} 位，共 ${activity.participantCount.toLocaleString()} 人报名`;

  // Winners section
  if (activity.status === 'drawed' && activity.drawResult) {
    const dr = activity.drawResult;
    document.getElementById('winnerSection').style.display = 'block';
    document.getElementById('winnerCards').innerHTML = dr.winners.map((w, i) => `
      <div class="winner-card">
        <div class="winner-rank">中奖 #${i + 1}</div>
        <div class="winner-index">${w.index}</div>
        <div class="winner-address">${w.address}</div>
      </div>
    `).join('');

    // Verify section
    document.getElementById('verifySection').style.display = 'block';
    document.getElementById('verifySeed').textContent = dr.seed;
    document.getElementById('verifyShanghai').textContent = dr.shanghaiIndex;
    document.getElementById('verifyShenzhen').textContent = dr.shenzhenIndex;
    document.getElementById('verifyBlock').textContent = dr.blockNumber.toLocaleString();
    document.getElementById('verifyBlockTs').textContent = dr.blockTs.toLocaleString();
    document.getElementById('verifyTx').textContent = dr.txHash.slice(0, 20) + '...';
  }
}

function handleRegister() {
  const user = getCurrentUser();
  if (!user) {
    window.location.href = 'login.html';
    return;
  }
  const params = new URLSearchParams(window.location.search);
  const activityId = params.get('id');
  showToast(`报名成功！活动 #${activityId}`, 'success');
  document.getElementById('registerBtn').disabled = true;
  document.getElementById('registerBtn').textContent = '已报名';
  document.getElementById('registerHint').textContent = '报名记录已上链，可在个人中心查看';
}

// ===== Verify Page =====

function doVerify(activityId, shanghai, shenzhen) {
  const activity = mockGetActivity(activityId);
  if (!activity || !activity.drawResult) {
    showToast('请选择已开奖的活动', 'error');
    return;
  }
  const raw = Math.floor(parseFloat(shanghai) * parseFloat(shenzhen) * 10000);
  const seed = raw.toString().split('').reverse().join('').slice(0, 12);
  return {
    calculated: seed,
    recorded: activity.drawResult.seed,
    match: seed === activity.drawResult.seed,
    raw,
  };
}
