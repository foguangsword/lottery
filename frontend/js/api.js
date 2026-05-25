// ===== Mock API =====
// 后端对接后替换为真实 fetch 调用

const delay = (ms = 200) => new Promise(r => setTimeout(r, ms));

// ===== Mock Data =====

const MOCK_ACTIVITIES = [
  {
    id: 1, name: '新年好运抽奖', contractName: 'new_year_2026',
    startTime: '2026-01-20 00:00', endTime: '2026-02-10 23:59',
    luckyCount: 5, status: 'drawed',
    participantCount: 2739, participants: [],
    drawResult: {
      seed: '579987024129',
      shanghaiIndex: '3250.18',
      shenzhenIndex: '10680.72',
      blockNumber: 20123456,
      blockTs: 1737561600,
      txHash: '0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234',
      winners: [
        { index: 12, address: '0x3f8c...a12b' },
        { index: 847, address: '0x7a2d...c34e' },
        { index: 1503, address: '0x9e4f...d56a' },
        { index: 2201, address: '0x1b6c...e78f' },
        { index: 2689, address: '0x5d3a...f901' },
      ],
    },
  },
  {
    id: 2, name: '春季特别抽奖', contractName: 'spring_special',
    startTime: '2026-03-01 00:00', endTime: '2026-05-31 23:59',
    luckyCount: 3, status: 'active',
    participantCount: 1562, participants: [],
  },
  {
    id: 3, name: '618 狂欢抽奖', contractName: 'june_promo',
    startTime: '2026-06-01 00:00', endTime: '2026-06-20 23:59',
    luckyCount: 10, status: 'active',
    participantCount: 428, participants: [],
  },
  {
    id: 4, name: '中秋团圆奖', contractName: 'mid_autumn_2025',
    startTime: '2025-09-01 00:00', endTime: '2025-09-25 23:59',
    luckyCount: 8, status: 'drawed',
    participantCount: 8921, participants: [],
    drawResult: {
      seed: '310248975630',
      shanghaiIndex: '3120.55',
      shenzhenIndex: '9890.30',
      blockNumber: 18987654,
      blockTs: 1727280000,
      txHash: '0xef567890ef567890ef567890ef567890ef567890ef567890ef567890ef567890',
      winners: [
        { index: 56, address: '0x21ab...b234' },
        { index: 1024, address: '0x43cd...c567' },
        { index: 3456, address: '0x65ef...d890' },
        { index: 5001, address: '0x87ab...e123' },
        { index: 6234, address: '0x09cd...f456' },
        { index: 7456, address: '0x12ef...a789' },
        { index: 8001, address: '0x34ab...b012' },
        { index: 8802, address: '0x56cd...c345' },
      ],
    },
  },
  {
    id: 5, name: '国庆黄金周抽奖', contractName: 'national_day',
    startTime: '2025-10-01 00:00', endTime: '2025-10-08 23:59',
    luckyCount: 20, status: 'canceled',
    participantCount: 499, participants: [],
  },
];

const MOCK_USER = {
  id: 1, username: 'demo_user', email: 'demo@example.com',
  ethAddress: '0x8f3c9e1a2b4d5f6c7e8d9a0b1c2d3e4f5a6b7c8d',
};

const MOCK_REGISTRATIONS = [
  { userId: 1, activityId: 1, index: 273, activityName: '新年好运抽奖', time: '2026-01-22 14:30', status: 'drawed', won: false },
  { userId: 1, activityId: 2, index: 512, activityName: '春季特别抽奖', time: '2026-03-15 09:20', status: 'active', won: false },
  { userId: 1, activityId: 4, index: 3456, activityName: '中秋团圆奖', time: '2025-09-10 16:45', status: 'drawed', won: true },
];

// ===== Mock Functions =====

function mockGetActivities() {
  return MOCK_ACTIVITIES.map(a => ({
    ...a,
    participants: Array.from({ length: Math.min(a.participantCount, 20) }, (_, i) => ({
      index: i, username: `用户${1000 + i}`, address: `0x${(i * 12345).toString(16).padStart(4, '0')}...${(i * 54321).toString(16).padStart(4, '0')}`,
      time: a.startTime,
    })),
  }));
}

function mockGetActivity(id) {
  const a = MOCK_ACTIVITIES.find(a => a.id == id);
  if (!a) return null;
  return {
    ...a,
    participants: Array.from({ length: Math.min(a.participantCount, 20) }, (_, i) => ({
      index: i, username: `用户${1000 + i}`, address: `0x${(i * 12345).toString(16).padStart(8, '0').slice(0, 8)}`,
      time: a.startTime,
    })),
  };
}

function mockLogin(email, password) {
  if (email && password.length >= 6) {
    localStorage.setItem('user', JSON.stringify(MOCK_USER));
    showToast('登录成功', 'success');
    return true;
  }
  showToast('邮箱或密码错误', 'error');
  return false;
}

function mockRegister(username, email, password) {
  if (username && email && password.length >= 6) {
    const addr = '0x' + Array.from({ length: 40 }, () => Math.floor(Math.random() * 16).toString(16)).join('');
    const user = { id: Date.now(), username, email, ethAddress: addr };
    localStorage.setItem('user', JSON.stringify(user));
    showToast('注册成功', 'success');
    return true;
  }
  return false;
}

function getCurrentUser() {
  const u = localStorage.getItem('user');
  return u ? JSON.parse(u) : null;
}

function mockGetUserRegistrations() {
  const user = getCurrentUser();
  if (!user) return [];
  return MOCK_REGISTRATIONS.filter(r => r.userId === user.id);
}

function mockGetUserWinnings() {
  const user = getCurrentUser();
  if (!user) return [];
  const regs = MOCK_REGISTRATIONS.filter(r => r.userId === user.id && r.won);
  return regs.map(r => {
    const activity = MOCK_ACTIVITIES.find(a => a.id === r.activityId);
    if (!activity || !activity.drawResult) return null;
    const winner = activity.drawResult.winners.find(w => w.index === r.index);
    return { ...r, winnerIndex: winner ? winner.index : r.index, activityId: r.activityId };
  }).filter(Boolean);
}
