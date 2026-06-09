<script setup>
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { ref, onMounted } from 'vue'

const router = useRouter();
const isAuthenticated = ref(false);

onMounted(() => {
  const token = localStorage.getItem("authToken");
  isAuthenticated.value = !!token; // ✅ Checks login status
});

const logout = () => {
  localStorage.clear();
  isAuthenticated.value = false;
  router.push("/login"); // ✅ Redirects to login
  window.location.reload(); // 🚀 **Forces UI update**
};
</script>

<template>
	<!-- WhatsApp Header -->
	<header class="whatsapp-header">
		<div class="header-content">
			<a class="brand" href="/">
				<svg class="brand-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>
				<span class="brand-name">WASAText</span>
			</a>
		</div>
	</header>

	<div class="app-container">
		<div class="app-layout">
			<!-- WhatsApp Sidebar -->
			<nav id="sidebarMenu" class="whatsapp-sidebar">
				<div class="sidebar-content">
					<!-- Navigation Section -->
					<div class="nav-section">
						<h6 class="sidebar-heading">
							<span>General</span>
						</h6>
						<ul class="nav-list">
							<li class="nav-item">
								<RouterLink to="/home" class="nav-link">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
										<polyline points="9 22 9 12 15 12 15 22"></polyline>
									</svg>
									<span class="nav-text">Home</span>
								</RouterLink>
							</li>
							<li class="nav-item">
								<RouterLink 
									to="/conversations" 
									class="nav-link" 
									:class="{ disabled: !isAuthenticated }">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
									</svg>
									<span class="nav-text">Chats</span>
								</RouterLink>
							</li>
							<!--
							<li class="nav-item">
								<RouterLink 
									to="/groups" 
									class="nav-link" 
									:class="{ disabled: !isAuthenticated }">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
										<circle cx="9" cy="7" r="4"></circle>
										<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
										<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
									</svg>
									<span class="nav-text">Groups</span>
								</RouterLink>
							</li>
							<li class="nav-item">
								<RouterLink 
									to="/users/me/username" 
									class="nav-link" 
									:class="{ disabled: !isAuthenticated }">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
										<circle cx="12" cy="7" r="4"></circle>
									</svg>
									<span class="nav-text">Profile</span>
								</RouterLink>
							</li>
							-->
						</ul>
					</div>

					<!-- Bottom Section (Login/Logout) -->
					<div class="nav-bottom">
						<ul class="nav-list">
							<li v-if="!isAuthenticated" class="nav-item">
								<RouterLink to="/login" class="nav-link login-link">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path>
										<polyline points="10 17 15 12 10 7"></polyline>
										<line x1="15" y1="12" x2="3" y2="12"></line>
									</svg>
									<span class="nav-text">Login</span>
								</RouterLink>
							</li>
							<li v-if="isAuthenticated" class="nav-item">
								<button class="nav-link logout-button" @click="logout">
									<svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
										<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
										<polyline points="16 17 21 12 16 7"></polyline>
										<line x1="21" y1="12" x2="9" y2="12"></line>
									</svg>
									<span class="nav-text">Logout</span>
								</button>
							</li>
						</ul>
					</div>
				</div>
			</nav>

			<!-- Main Content Area -->
			<main class="whatsapp-main">
				<RouterView />
			</main>
		</div>
	</div>
</template>

<style>
/* WhatsApp Color Palette */
:root {
	--wa-bg-primary: #111b21;
	--wa-bg-secondary: #202c33;
	--wa-bg-hover: #2a3942;
	--wa-border: #2a3942;
	--wa-text-primary: #e9edef;
	--wa-text-secondary: #aebac1;
	--wa-text-muted: #667781;
	--wa-accent: #00a884;
	--wa-accent-hover: #06cf9c;
	--wa-danger: #ea4335;
}

/* Reset & Global Styles */
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	background-color: var(--wa-bg-primary);
	color: var(--wa-text-primary);
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* Header Styles */
.whatsapp-header {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	height: 60px;
	background-color: var(--wa-bg-secondary);
	border-bottom: 1px solid var(--wa-border);
	z-index: 1000;
	display: flex;
	align-items: center;
	padding: 0 20px;
}

.header-content {
	display: flex;
	align-items: center;
	width: 100%;
}

.brand {
	display: flex;
	align-items: center;
	text-decoration: none;
	color: var(--wa-text-primary);
	transition: color 0.2s ease;
}

.brand:hover {
	color: var(--wa-accent);
}

.brand-icon {
	width: 32px;
	height: 32px;
	margin-right: 12px;
	stroke-width: 2;
	stroke-linecap: round;
	stroke-linejoin: round;
}

.brand-name {
	font-size: 1.25rem;
	font-weight: 600;
}

/* App Container */
.app-container {
	margin-top: 60px;
	height: calc(100vh - 60px);
}

.app-layout {
	display: flex;
	height: 100%;
}

/* Sidebar Styles */
.whatsapp-sidebar {
	width: 280px;
	background-color: var(--wa-bg-primary);
	border-right: 1px solid var(--wa-border);
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

.sidebar-content {
	display: flex;
	flex-direction: column;
	height: 100%;
	padding: 12px 0;
}

.nav-section {
	flex: 1;
	overflow-y: auto;
	padding: 0 8px;
}

.sidebar-heading {
	font-size: 0.75rem;
	color: var(--wa-text-muted);
	padding: 12px 16px 8px;
	text-transform: uppercase;
	letter-spacing: 0.5px;
	font-weight: 500;
}

.nav-bottom {
	padding: 8px;
	border-top: 1px solid var(--wa-border);
}

/* Navigation List */
.nav-list {
	list-style: none;
	margin: 0;
	padding: 0;
}

.nav-item {
	margin: 2px 0;
}

/* Navigation Links */
.nav-link {
	display: flex;
	align-items: center;
	padding: 12px 16px;
	color: var(--wa-text-secondary);
	text-decoration: none;
	border-radius: 8px;
	transition: all 0.2s ease;
	cursor: pointer;
	border: none;
	background: none;
	width: 100%;
	text-align: left;
	font-size: 0.95rem;
	position: relative;
}

.nav-link:hover {
	background-color: var(--wa-bg-hover);
	color: var(--wa-text-primary);
}

.nav-link.router-link-active {
	background-color: var(--wa-bg-hover);
	color: var(--wa-accent);
}

.nav-link.router-link-active .nav-icon {
	color: var(--wa-accent);
}

/* Navigation Icons */
.nav-icon {
	width: 20px;
	height: 20px;
	margin-right: 16px;
	stroke-width: 2;
	stroke-linecap: round;
	stroke-linejoin: round;
	flex-shrink: 0;
	transition: color 0.2s ease;
}

.nav-text {
	flex: 1;
}

/* Disabled State */
.nav-link.disabled {
	pointer-events: none;
	opacity: 0.4;
	cursor: not-allowed;
}

/* Logout Button */
.logout-button {
	color: var(--wa-danger);
}

.logout-button:hover {
	background-color: rgba(234, 67, 53, 0.1);
	color: var(--wa-danger);
}

.logout-button .nav-icon {
	color: var(--wa-danger);
}

/* Login Link */
.login-link {
	color: var(--wa-accent);
}

.login-link:hover {
	background-color: rgba(0, 168, 132, 0.1);
	color: var(--wa-accent-hover);
}

.login-link .nav-icon {
	color: var(--wa-accent);
}

/* Main Content Area */
.whatsapp-main {
	flex: 1;
	background-color: var(--wa-bg-primary);
	overflow-y: auto;
	position: relative;
}

/* Scrollbar Styles */
.nav-section::-webkit-scrollbar,
.whatsapp-main::-webkit-scrollbar {
	width: 6px;
}

.nav-section::-webkit-scrollbar-track,
.whatsapp-main::-webkit-scrollbar-track {
	background: var(--wa-bg-primary);
}

.nav-section::-webkit-scrollbar-thumb,
.whatsapp-main::-webkit-scrollbar-thumb {
	background: #374955;
	border-radius: 3px;
}

.nav-section::-webkit-scrollbar-thumb:hover,
.whatsapp-main::-webkit-scrollbar-thumb:hover {
	background: #4a5c6a;
}

/* Responsive Design */
@media (max-width: 768px) {
	.whatsapp-sidebar {
		position: fixed;
		left: -280px;
		top: 60px;
		height: calc(100vh - 60px);
		z-index: 999;
		transition: left 0.3s ease;
	}

	.whatsapp-sidebar.open {
		left: 0;
	}

	.whatsapp-main {
		width: 100%;
	}

	.brand-name {
		font-size: 1rem;
	}
}

@media (max-width: 480px) {
	.whatsapp-sidebar {
		width: 100%;
		left: -100%;
	}

	.brand-icon {
		width: 28px;
		height: 28px;
	}
}
</style>

