// ========== 交互功能 ==========

document.addEventListener('DOMContentLoaded', function() {
    // 导航栏滚动效果
    const navbar = document.querySelector('.navbar');
    let lastScrollY = window.scrollY;

    window.addEventListener('scroll', () => {
        const currentScrollY = window.scrollY;

        if (currentScrollY > 100) {
            navbar.style.background = 'rgba(255, 255, 255, 0.98)';
            navbar.style.boxShadow = '0 4px 6px -1px rgba(0, 0, 0, 0.1)';
        } else {
            navbar.style.background = 'rgba(255, 255, 255, 0.95)';
            navbar.style.boxShadow = '0 1px 2px 0 rgba(0, 0, 0, 0.05)';
        }

        // 隐藏/显示导航栏
        if (currentScrollY > lastScrollY && currentScrollY > 200) {
            navbar.style.transform = 'translateY(-100%)';
        } else {
            navbar.style.transform = 'translateY(0)';
        }

        lastScrollY = currentScrollY;
    });

    // 移动端菜单切换
    const navToggle = document.querySelector('.nav-toggle');
    const navMenu = document.querySelector('.nav-menu');

    if (navToggle) {
        navToggle.addEventListener('click', () => {
            navMenu.classList.toggle('active');
        });
    }

    // 平滑滚动
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // 按钮点击效果
    document.querySelectorAll('button').forEach(button => {
        button.addEventListener('click', function(e) {
            // 创建涟漪效果
            const ripple = document.createElement('span');
            const rect = this.getBoundingClientRect();
            const size = Math.max(rect.width, rect.height);
            const x = e.clientX - rect.left - size / 2;
            const y = e.clientY - rect.top - size / 2;

            ripple.style.width = ripple.style.height = size + 'px';
            ripple.style.left = x + 'px';
            ripple.style.top = y + 'px';
            ripple.classList.add('ripple');

            this.appendChild(ripple);

            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });

    // 统计数字动画
    function animateNumbers() {
        const statNumbers = document.querySelectorAll('.stat-number, .stat-value');

        statNumbers.forEach(stat => {
            const text = stat.textContent;
            const number = parseInt(text.replace(/[^\d]/g, ''));

            if (isNaN(number)) return;

            let current = 0;
            const increment = number / 50;
            const duration = 2000; // 2秒
            const stepTime = duration / 50;

            const timer = setInterval(() => {
                current += increment;
                if (current >= number) {
                    stat.textContent = text;
                    clearInterval(timer);
                } else {
                    if (text.includes('%')) {
                        stat.textContent = Math.floor(current) + '%';
                    } else if (text.includes('+')) {
                        stat.textContent = Math.floor(current) + '+';
                    } else {
                        stat.textContent = Math.floor(current);
                    }
                }
            }, stepTime);
        });
    }

    // 观察器 - 当元素进入视图时触发动画
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate');

                // 如果是统计区域，触发动画数字
                if (entry.target.classList.contains('stats') ||
                    entry.target.classList.contains('hero')) {
                    animateNumbers();
                }
            }
        });
    }, {
        threshold: 0.1
    });

    // 观察所有卡片和区域
    document.querySelectorAll('.feature-card, .module-card, .stat-card, .hero, .stats').forEach(el => {
        observer.observe(el);
    });

    // 登录按钮点击事件
    const loginBtn = document.querySelector('.btn-login');
    if (loginBtn) {
        loginBtn.addEventListener('click', () => {
            showNotification('登录功能开发中...', 'info');
        });
    }

    // 开始分享按钮
    const shareBtn = document.querySelector('.btn-primary');
    if (shareBtn && shareBtn.textContent.includes('开始分享')) {
        shareBtn.addEventListener('click', () => {
            showNotification('感谢您的关注！功能即将上线', 'success');
        });
    }

    // API文档按钮
    const apiDocBtns = document.querySelectorAll('.api-actions button');
    apiDocBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            if (btn.textContent.includes('API文档')) {
                showNotification('API文档功能开发中...', 'info');
            } else if (btn.textContent.includes('Postman')) {
                showNotification('Postman集合即将提供', 'info');
            }
        });
    });

    // 通知函数
    function showNotification(message, type = 'info') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <i class="fas ${type === 'success' ? 'fa-check-circle' : 'fa-info-circle'}"></i>
                <span>${message}</span>
            </div>
            <button class="notification-close">
                <i class="fas fa-times"></i>
            </button>
        `;

        // 添加样式
        notification.style.cssText = `
            position: fixed;
            top: 100px;
            right: 20px;
            background: white;
            padding: 15px 20px;
            border-radius: 10px;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
            display: flex;
            align-items: center;
            gap: 15px;
            z-index: 10000;
            animation: slideIn 0.3s ease;
            max-width: 400px;
            border-left: 4px solid ${type === 'success' ? '#10b981' : '#3b82f6'};
        `;

        // 添加到页面
        document.body.appendChild(notification);

        // 关闭按钮事件
        const closeBtn = notification.querySelector('.notification-close');
        closeBtn.addEventListener('click', () => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        });

        // 自动关闭
        setTimeout(() => {
            if (notification.parentNode) {
                notification.style.animation = 'slideOut 0.3s ease';
                setTimeout(() => notification.remove(), 300);
            }
        }, 3000);
    }

    // 添加通知动画CSS
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from {
                transform: translateX(100%);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }

        @keyframes slideOut {
            from {
                transform: translateX(0);
                opacity: 1;
            }
            to {
                transform: translateX(100%);
                opacity: 0;
            }
        }

        .notification-content {
            display: flex;
            align-items: center;
            gap: 10px;
            flex: 1;
        }

        .notification i {
            font-size: 1.2rem;
        }

        .notification.success i {
            color: #10b981;
        }

        .notification.info i {
            color: #3b82f6;
        }

        .notification-close {
            background: none;
            border: none;
            cursor: pointer;
            font-size: 1rem;
            color: #6b7280;
            padding: 5px;
            transition: color 0.3s ease;
        }

        .notification-close:hover {
            color: #374151;
        }

        .ripple {
            position: absolute;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.6);
            transform: scale(0);
            animation: rippleEffect 0.6s linear;
            pointer-events: none;
        }

        @keyframes rippleEffect {
            to {
                transform: scale(4);
                opacity: 0;
            }
        }
    `;
    document.head.appendChild(style);

    // 页面加载完成后的欢迎消息
    setTimeout(() => {
        showNotification('🎉 欢迎来到资源分享平台！', 'success');
    }, 1000);
});

// ========== 工具函数 ==========

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 节流函数
function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// 获取元素在页面中的位置
function getOffset(element) {
    let offsetTop = 0;
    let offsetLeft = 0;

    while (element) {
        offsetTop += element.offsetTop;
        offsetLeft += element.offsetLeft;
        element = element.offsetParent;
    }

    return {
        top: offsetTop,
        left: offsetLeft
    };
}
