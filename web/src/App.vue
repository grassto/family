<template>
  <div id="app">
    <header class="header">
      <div class="header-inner">
        <h1 class="logo">🌳 家族图谱</h1>
        <nav class="nav">
          <router-link to="/families">家族管理</router-link>
          <router-link to="/birthdays">生日提醒</router-link>
        </nav>
      </div>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  background: #f5f7fa;
  color: #333;
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
}

.logo {
  font-size: 20px;
  font-weight: 600;
}

.nav a {
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
  margin-left: 24px;
  font-size: 15px;
  transition: color 0.2s;
}

.nav a:hover,
.nav a.router-link-active {
  color: #fff;
}

.main {
  max-width: 1200px;
  margin: 24px auto;
  padding: 0 24px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5a6fd6;
}

.btn-danger {
  background: #ff4757;
  color: white;
}

.btn-danger:hover {
  background: #e8404f;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 13px;
}

.card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 16px;
  padding: 28px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal h3 {
  margin-bottom: 20px;
  font-size: 18px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.empty {
  text-align: center;
  padding: 60px 20px;
  color: #999;
  font-size: 15px;
}

.tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.tag-male {
  background: #e8f4fd;
  color: #1890ff;
}

.tag-female {
  background: #fff0f6;
  color: #eb2f96;
}

.tag-alive {
  background: #f6ffed;
  color: #52c41a;
}

.tag-dead {
  background: #f5f5f5;
  color: #999;
}

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .header {
    padding: 0 16px;
  }

  .header-inner {
    height: 52px;
  }

  .logo {
    font-size: 17px;
  }

  .nav a {
    margin-left: 16px;
    font-size: 14px;
  }

  .main {
    margin: 16px auto;
    padding: 0 16px;
  }

  .card {
    padding: 16px;
    border-radius: 10px;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }

  /* 弹窗全屏 */
  .modal-overlay {
    align-items: flex-end;
  }

  .modal {
    width: 100%;
    max-width: 100%;
    border-radius: 16px 16px 0 0;
    max-height: 90vh;
    padding: 20px;
  }
}
</style>
