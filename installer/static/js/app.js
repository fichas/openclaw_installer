// OpenClaw 配置向导 - 前端逻辑

(function() {
    'use strict';

    // 验证规则
    const validators = {
        required: (value) => value && value.trim() !== '',
        minLength: (value, length) => value && value.length >= length,
        maxLength: (value, length) => !value || value.length <= length,
        port: (value) => {
            const port = parseInt(value, 10);
            return !isNaN(port) && port >= 1 && port <= 65535;
        },
        url: (value) => {
            if (!value) return true; // 可选字段
            try {
                new URL(value);
                return true;
            } catch {
                return false;
            }
        },
        corpId: (value) => /^ww[a-f0-9]{16,18}$/i.test(value),
        agentId: (value) => /^\d+$/.test(value),
        appKey: (value) => /^ding[a-z0-9]{16,20}$/i.test(value),
        appId: (value) => /^cli_[a-z0-9]{16}$/i.test(value),
    };

    // 显示/隐藏错误
    function showError(fieldId, message) {
        const field = document.getElementById(fieldId);
        const errorEl = document.getElementById(fieldId + 'Error');
        if (field) field.classList.add('error');
        if (errorEl) errorEl.textContent = message;
    }

    function clearError(fieldId) {
        const field = document.getElementById(fieldId);
        const errorEl = document.getElementById(fieldId + 'Error');
        if (field) field.classList.remove('error');
        if (errorEl) errorEl.textContent = '';
    }

    function clearAllErrors(form) {
        form.querySelectorAll('.error').forEach(el => el.classList.remove('error'));
        form.querySelectorAll('.error-message').forEach(el => el.textContent = '');
    }

    // 基础配置表单验证
    function validateBasicConfig(form) {
        let isValid = true;
        clearAllErrors(form);

        // AI 提供商
        const aiProvider = form.querySelector('#aiProvider');
        if (!validators.required(aiProvider.value)) {
            showError('aiProvider', '请选择 AI 提供商');
            isValid = false;
        }

        // AI 模型
        const aiModel = form.querySelector('#aiModel');
        if (!validators.required(aiModel.value)) {
            showError('aiModel', '请输入模型名称');
            isValid = false;
        }

        // API Key
        const aiApiKey = form.querySelector('#aiApiKey');
        if (!validators.required(aiApiKey.value)) {
            showError('aiApiKey', '请输入 API Key');
            isValid = false;
        }

        // API Base URL
        const aiBaseUrl = form.querySelector('#aiBaseUrl');
        if (aiBaseUrl.value && !validators.url(aiBaseUrl.value)) {
            showError('aiBaseUrl', '请输入有效的 URL');
            isValid = false;
        }

        // 服务端口
        const serverPort = form.querySelector('#serverPort');
        if (!validators.port(serverPort.value)) {
            showError('serverPort', '请输入有效的端口号（1-65535）');
            isValid = false;
        }

        return isValid;
    }

    // 平台配置表单验证
    function validatePlatformConfig(form) {
        let isValid = true;
        let hasEnabledPlatform = false;
        clearAllErrors(form);

        const wecomEnabled = form.querySelector('#wecomEnabled').checked;
        const dingtalkEnabled = form.querySelector('#dingtalkEnabled').checked;
        const larkEnabled = form.querySelector('#larkEnabled').checked;

        // 企业微信验证
        if (wecomEnabled) {
            hasEnabledPlatform = true;

            const corpID = form.querySelector('#wecomCorpID');
            if (!validators.required(corpID.value)) {
                showError('wecomCorpID', '请输入企业 ID');
                isValid = false;
            } else if (!validators.corpId(corpID.value)) {
                showError('wecomCorpID', '企业 ID 格式不正确（应以 ww 开头）');
                isValid = false;
            }

            const agentID = form.querySelector('#wecomAgentID');
            if (!validators.required(agentID.value)) {
                showError('wecomAgentID', '请输入 AgentID');
                isValid = false;
            } else if (!validators.agentId(agentID.value)) {
                showError('wecomAgentID', 'AgentID 应为数字');
                isValid = false;
            }

            const secret = form.querySelector('#wecomSecret');
            if (!validators.required(secret.value)) {
                showError('wecomSecret', '请输入应用 Secret');
                isValid = false;
            }
        }

        // 钉钉验证
        if (dingtalkEnabled) {
            hasEnabledPlatform = true;

            const appKey = form.querySelector('#dingtalkAppKey');
            if (!validators.required(appKey.value)) {
                showError('dingtalkAppKey', '请输入 AppKey');
                isValid = false;
            } else if (!validators.appKey(appKey.value)) {
                showError('dingtalkAppKey', 'AppKey 格式不正确（应以 ding 开头）');
                isValid = false;
            }

            const appSecret = form.querySelector('#dingtalkAppSecret');
            if (!validators.required(appSecret.value)) {
                showError('dingtalkAppSecret', '请输入 AppSecret');
                isValid = false;
            }
        }

        // 飞书验证
        if (larkEnabled) {
            hasEnabledPlatform = true;

            const appID = form.querySelector('#larkAppID');
            if (!validators.required(appID.value)) {
                showError('larkAppID', '请输入 App ID');
                isValid = false;
            } else if (!validators.appId(appID.value)) {
                showError('larkAppID', 'App ID 格式不正确（应以 cli_ 开头）');
                isValid = false;
            }

            const appSecret = form.querySelector('#larkAppSecret');
            if (!validators.required(appSecret.value)) {
                showError('larkAppSecret', '请输入 App Secret');
                isValid = false;
            }
        }

        // 检查是否至少启用了一个平台
        const platformError = document.getElementById('platformError');
        if (!hasEnabledPlatform) {
            platformError.classList.add('show');
            isValid = false;
        } else {
            platformError.classList.remove('show');
        }

        return isValid;
    }

    // 保存配置到后端
    async function saveConfig(config) {
        const response = await fetch('/api/config', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(config),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || '保存配置失败');
        }

        return response.json();
    }

    // 获取基础配置数据
    function getBasicConfigData(form) {
        return {
            aiProvider: form.querySelector('#aiProvider').value,
            aiModel: form.querySelector('#aiModel').value,
            aiApiKey: form.querySelector('#aiApiKey').value,
            aiBaseUrl: form.querySelector('#aiBaseUrl').value,
            serverPort: parseInt(form.querySelector('#serverPort').value, 10),
            logLevel: form.querySelector('#logLevel').value,
        };
    }

    // 获取平台配置数据
    function getPlatformConfigData(form) {
        return {
            wecomEnabled: form.querySelector('#wecomEnabled').checked,
            wecomCorpID: form.querySelector('#wecomCorpID').value,
            wecomAgentID: form.querySelector('#wecomAgentID').value,
            wecomSecret: form.querySelector('#wecomSecret').value,
            wecomToken: form.querySelector('#wecomToken').value,
            wecomEncodingAESKey: form.querySelector('#wecomEncodingAESKey').value,
            dingtalkEnabled: form.querySelector('#dingtalkEnabled').checked,
            dingtalkAppKey: form.querySelector('#dingtalkAppKey').value,
            dingtalkAppSecret: form.querySelector('#dingtalkAppSecret').value,
            dingtalkRobotCode: form.querySelector('#dingtalkRobotCode').value,
            larkEnabled: form.querySelector('#larkEnabled').checked,
            larkAppID: form.querySelector('#larkAppID').value,
            larkAppSecret: form.querySelector('#larkAppSecret').value,
            larkEncryptKey: form.querySelector('#larkEncryptKey').value,
            larkVerificationToken: form.querySelector('#larkVerificationToken').value,
        };
    }

    // 初始化平台标签切换
    function initPlatformTabs() {
        const tabs = document.querySelectorAll('.platform-tab');
        const platformError = document.getElementById('platformError');

        tabs.forEach(tab => {
            const platform = tab.dataset.platform;
            const checkbox = tab.querySelector('input[type="checkbox"]');
            const configPanel = document.getElementById(platform + 'Config');

            // 标签点击切换激活状态
            tab.addEventListener('click', (e) => {
                if (e.target.type !== 'checkbox') {
                    tabs.forEach(t => t.classList.remove('active'));
                    tab.classList.add('active');

                    // 显示对应的配置面板
                    document.querySelectorAll('.platform-config').forEach(panel => {
                        panel.style.display = 'none';
                    });
                    if (configPanel) {
                        configPanel.style.display = 'block';
                    }
                }
            });

            // 复选框切换启用状态
            if (checkbox) {
                checkbox.addEventListener('change', () => {
                    tab.classList.toggle('active', checkbox.checked);
                    if (configPanel) {
                        configPanel.style.display = checkbox.checked ? 'block' : 'none';
                    }
                    if (platformError) {
                        platformError.classList.remove('show');
                    }
                });
            }
        });
    }

    // 初始化基础配置表单
    function initBasicConfigForm() {
        const form = document.getElementById('basicConfigForm');
        if (!form) return;

        // 实时验证
        form.querySelectorAll('input, select').forEach(field => {
            field.addEventListener('blur', () => {
                clearError(field.id);
            });

            field.addEventListener('input', () => {
                if (field.classList.contains('error')) {
                    clearError(field.id);
                }
            });
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            if (!validateBasicConfig(form)) {
                return;
            }

            const submitBtn = form.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.classList.add('loading');
            submitBtn.textContent = '保存中...';

            try {
                const config = getBasicConfigData(form);
                await saveConfig({ type: 'basic', data: config });

                // 跳转到下一步
                const nextUrl = form.dataset.next;
                if (nextUrl) {
                    window.location.href = nextUrl;
                }
            } catch (error) {
                alert('保存失败：' + error.message);
                submitBtn.disabled = false;
                submitBtn.classList.remove('loading');
                submitBtn.textContent = originalText;
            }
        });
    }

    // 初始化平台配置表单
    function initPlatformConfigForm() {
        const form = document.getElementById('platformConfigForm');
        if (!form) return;

        initPlatformTabs();

        // 实时验证
        form.querySelectorAll('input').forEach(field => {
            field.addEventListener('blur', () => {
                clearError(field.id);
            });

            field.addEventListener('input', () => {
                if (field.classList.contains('error')) {
                    clearError(field.id);
                }
            });
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            if (!validatePlatformConfig(form)) {
                return;
            }

            const submitBtn = form.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.classList.add('loading');
            submitBtn.textContent = '保存中...';

            try {
                const config = getPlatformConfigData(form);
                await saveConfig({ type: 'platform', data: config });

                // 跳转到完成页
                const nextUrl = form.dataset.next;
                if (nextUrl) {
                    window.location.href = nextUrl;
                }
            } catch (error) {
                alert('保存失败：' + error.message);
                submitBtn.disabled = false;
                submitBtn.classList.remove('loading');
                submitBtn.textContent = originalText;
            }
        });
    }

    // 启动服务
    window.startService = async function() {
        const btn = document.querySelector('.finish-card .btn-primary');
        if (!btn) return;

        const originalText = btn.textContent;
        btn.disabled = true;
        btn.classList.add('loading');
        btn.textContent = '启动中...';

        try {
            const response = await fetch('/api/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || '启动服务失败');
            }

            alert('服务启动成功！');
            btn.textContent = '已启动';
        } catch (error) {
            alert('启动失败：' + error.message);
            btn.disabled = false;
            btn.classList.remove('loading');
            btn.textContent = originalText;
        }
    };

    // 初始化
    document.addEventListener('DOMContentLoaded', () => {
        initBasicConfigForm();
        initPlatformConfigForm();
    });
})();
