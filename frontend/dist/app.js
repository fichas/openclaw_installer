// OpenClaw 安装向导 - 前端逻辑

// 当前步骤
let currentStep = 1;
const totalSteps = 4;

// 适配器配置状态
const adapterState = {
    wecom: false,
    dingtalk: false,
    feishu: false
};

// 安装进度模拟
let installProgress = 0;
let progressInterval = null;

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    initEventListeners();
    updateStepIndicator();
});

// 初始化事件监听
function initEventListeners() {
    // 键盘导航
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && e.ctrlKey) {
            nextStep();
        }
    });
}

// 下一步
function nextStep() {
    if (currentStep >= totalSteps) return;

    const currentSection = document.getElementById(`step-${currentStep}`);
    const nextSection = document.getElementById(`step-${currentStep + 1}`);

    // 验证当前步骤
    if (!validateStep(currentStep)) {
        return;
    }

    // 动画切换
    currentSection.classList.remove('active');
    currentSection.classList.add('prev');

    setTimeout(() => {
        currentSection.classList.remove('prev');
        nextSection.classList.add('active');
    }, 100);

    currentStep++;
    updateStepIndicator();

    // 特殊步骤处理
    if (currentStep === 3) {
        updateConfigSections();
    } else if (currentStep === 4) {
        startInstall();
    }
}

// 上一步
function prevStep() {
    if (currentStep <= 1) return;

    const currentSection = document.getElementById(`step-${currentStep}`);
    const prevSection = document.getElementById(`step-${currentStep - 1}`);

    currentSection.classList.remove('active');

    setTimeout(() => {
        prevSection.classList.add('active');
    }, 50);

    currentStep--;
    updateStepIndicator();

    // 如果回到第4步且安装已完成，显示完成界面
    if (currentStep === 4) {
        const progressContainer = document.getElementById('install-progress');
        const completeContainer = document.getElementById('install-complete');

        if (completeContainer.style.display === 'flex') {
            progressContainer.style.display = 'none';
            completeContainer.style.display = 'flex';
        }
    }
}

// 更新步骤指示器
function updateStepIndicator() {
    const steps = document.querySelectorAll('.step-item');
    steps.forEach((step, index) => {
        const stepNum = index + 1;
        step.classList.remove('active', 'completed');

        if (stepNum === currentStep) {
            step.classList.add('active');
        } else if (stepNum < currentStep) {
            step.classList.add('completed');
        }
    });
}

// 验证步骤
function validateStep(step) {
    switch (step) {
        case 1:
            return true; // 欢迎页无需验证

        case 2:
            // 至少选择核心组件
            return true;

        case 3:
            // 验证 AI 配置
            const provider = document.getElementById('ai-provider').value;
            const apiKey = document.getElementById('api-key').value;

            if (!provider) {
                showError('请选择 AI 模型提供商');
                return false;
            }
            if (!apiKey) {
                showError('请输入 API 密钥');
                return false;
            }

            // 验证已启用的适配器配置
            if (adapterState.wecom) {
                const corpid = document.getElementById('wecom-corpid').value;
                const agentid = document.getElementById('wecom-agentid').value;
                const secret = document.getElementById('wecom-secret').value;
                if (!corpid || !agentid || !secret) {
                    showError('请完整填写企业微信配置');
                    return false;
                }
            }

            if (adapterState.dingtalk) {
                const appkey = document.getElementById('dingtalk-appkey').value;
                const appsecret = document.getElementById('dingtalk-appsecret').value;
                if (!appkey || !appsecret) {
                    showError('请完整填写钉钉配置');
                    return false;
                }
            }

            if (adapterState.feishu) {
                const appid = document.getElementById('feishu-appid').value;
                const appsecret = document.getElementById('feishu-appsecret').value;
                if (!appid || !appsecret) {
                    showError('请完整填写飞书配置');
                    return false;
                }
            }

            return true;

        default:
            return true;
    }
}

// 切换适配器选择
function toggleAdapter(adapter) {
    adapterState[adapter] = !adapterState[adapter];

    const checkbox = document.getElementById(`${adapter}-check`);
    const card = document.getElementById(`${adapter}-option`);

    checkbox.checked = adapterState[adapter];
    card.classList.toggle('selected', adapterState[adapter]);
}

// 更新配置区域显示
function updateConfigSections() {
    document.getElementById('wecom-config').style.display =
        adapterState.wecom ? 'block' : 'none';
    document.getElementById('dingtalk-config').style.display =
        adapterState.dingtalk ? 'block' : 'none';
    document.getElementById('feishu-config').style.display =
        adapterState.feishu ? 'block' : 'none';
}

// 开始安装
function startInstall() {
    const progressContainer = document.getElementById('install-progress');
    const completeContainer = document.getElementById('install-complete');

    progressContainer.style.display = 'flex';
    completeContainer.style.display = 'none';

    installProgress = 0;
    updateProgress(0);

    // 模拟安装进度
    const steps = [
        { progress: 10, status: '准备安装环境...', delay: 500 },
        { progress: 25, status: '安装 OpenClaw 核心组件...', delay: 1000 },
        { progress: 40, status: '配置 AI 服务连接...', delay: 800 },
        { progress: 55, status: adapterState.wecom ? '安装企业微信适配器...' : '跳过企业微信适配器...', delay: 1000 },
        { progress: 70, status: adapterState.dingtalk ? '安装钉钉适配器...' : '跳过钉钉适配器...', delay: 1000 },
        { progress: 85, status: adapterState.feishu ? '安装飞书适配器...' : '跳过飞书适配器...', delay: 1000 },
        { progress: 95, status: '生成配置文件...', delay: 800 },
        { progress: 100, status: '安装完成！', delay: 500 }
    ];

    let currentIndex = 0;

    function processStep() {
        if (currentIndex >= steps.length) {
            setTimeout(showComplete, 500);
            return;
        }

        const step = steps[currentIndex];
        updateProgress(step.progress, step.status);
        addLog(step.status);

        currentIndex++;
        setTimeout(processStep, step.delay);
    }

    processStep();
}

// 更新进度条
function updateProgress(progress, status) {
    installProgress = progress;

    const fill = document.getElementById('progress-fill');
    const text = document.getElementById('progress-text');

    fill.style.width = `${progress}%`;
    text.textContent = `${progress}%`;

    // 更新步骤状态
    const steps = document.querySelectorAll('.pstep');
    const activeIndex = Math.floor(progress / 20);

    steps.forEach((step, index) => {
        step.classList.remove('active', 'completed');
        if (index < activeIndex) {
            step.classList.add('completed');
            step.querySelector('.pstep-icon').textContent = '✓';
        } else if (index === activeIndex) {
            step.classList.add('active');
            step.querySelector('.pstep-icon').textContent = '●';
        } else {
            step.querySelector('.pstep-icon').textContent = '○';
        }
    });
}

// 添加日志
function addLog(message, type = 'info') {
    const logContainer = document.getElementById('progress-log');
    const line = document.createElement('div');
    line.className = `log-line ${type}`;
    line.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
    logContainer.appendChild(line);
    logContainer.scrollTop = logContainer.scrollHeight;
}

// 显示完成界面
function showComplete() {
    const progressContainer = document.getElementById('install-progress');
    const completeContainer = document.getElementById('install-complete');

    progressContainer.style.display = 'none';
    completeContainer.style.display = 'flex';

    // 更新安装摘要
    const adapters = [];
    adapters.push('核心');
    if (adapterState.wecom) adapters.push('企业微信');
    if (adapterState.dingtalk) adapters.push('钉钉');
    if (adapterState.feishu) adapters.push('飞书');

    document.getElementById('installed-adapters').textContent = adapters.join('、');

    const port = document.getElementById('gateway-port').value || '8080';
    document.getElementById('gateway-url').textContent = `http://localhost:${port}`;
}

// 启动服务
function launchService() {
    const autoStart = document.getElementById('auto-start').checked;
    const minimizeTray = document.getElementById('minimize-tray').checked;

    addLog(`启动服务... (开机启动: ${autoStart}, 最小化到托盘: ${minimizeTray})`);

    // 调用后端启动服务
    if (window.go && window.go.main && window.go.main.App) {
        window.go.main.App.LaunchApp()
            .then(() => {
                addLog('服务启动成功！', 'success');
                showNotification('OpenClaw 服务已启动', 'success');
            })
            .catch(err => {
                addLog(`服务启动失败: ${err}`, 'error');
                showError(`服务启动失败: ${err}`);
            });
    } else {
        // 开发环境模拟
        setTimeout(() => {
            addLog('服务启动成功！', 'success');
            showNotification('OpenClaw 服务已启动', 'success');
        }, 1000);
    }
}

// 显示错误
function showError(message) {
    // 简单的错误提示，可以扩展为更美观的弹窗
    const toast = document.createElement('div');
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: #f48771;
        color: white;
        padding: 12px 20px;
        border-radius: 6px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 1000;
        font-size: 14px;
        animation: slideIn 0.3s ease;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// 显示通知
function showNotification(message, type = 'info') {
    const toast = document.createElement('div');
    const bgColor = type === 'success' ? '#4ec9b0' : '#75beff';
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${bgColor};
        color: #1e1e1e;
        padding: 12px 20px;
        border-radius: 6px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 1000;
        font-size: 14px;
        font-weight: 500;
        animation: slideIn 0.3s ease;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Wails 运行时事件监听（如果可用）
if (window.runtime) {
    window.runtime.EventsOn('install:progress', (data) => {
        updateProgress(data.progress, data.status);
        addLog(data.status);
    });

    window.runtime.EventsOn('install:complete', (data) => {
        showComplete();
    });

    window.runtime.EventsOn('install:error', (data) => {
        addLog(`错误: ${data.error}`, 'error');
        showError(data.error);
    });
}

// 导出函数供全局使用
window.nextStep = nextStep;
window.prevStep = prevStep;
window.toggleAdapter = toggleAdapter;
window.startInstall = startInstall;
window.launchService = launchService;
