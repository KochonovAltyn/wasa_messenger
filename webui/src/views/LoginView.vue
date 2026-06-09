<template>
  <div class="login-container">
    <h1>Login</h1>
    <form @submit.prevent="doLogin">
      <div class="form-group">
        <label for="username">Username</label>
        <input
          type="text"
          id="username"
          v-model="username"
          placeholder="Enter your username"
          required
        />
      </div>
      <button type="submit">Login</button>
    </form>
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "LoginView",
  data() {
    return {
      username: "",
      errorMessage: "",
    };
  },
  methods: {
    async doLogin() {
      try {
        const response = await axios.post("/session", {
          username: this.username,
        });
        if (response.status === 201) {
          alert("User created successfully!");
        } else if (response.status === 200) {
          alert("Login successful!");
        }
        if (response.data.token && response.data.user) {
          localStorage.setItem("authToken", response.data.token);
          localStorage.setItem("username", response.data.user.username);
          localStorage.setItem("userID", response.data.user.id);
          // ✅ Restore profile photo from localStorage if it exists
          const profilePhoto = localStorage.getItem(`profilePhoto_${response.data.user.id}`);
          if (profilePhoto) {
            localStorage.setItem("profilePhoto", profilePhoto);
          }
        } else {
          throw new Error("Invalid login response");
        }
        // ✅ FORCE UI UPDATE & REDIRECT
        this.$router.push("/conversations").then(() => {
          window.location.reload(); // 🔄 Forces UI refresh
        });
      } catch (error) {
        this.errorMessage = error.response?.data?.error || "An error occurred.";
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.login-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 60px 20px 40px;
  color: #e9edef;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.login-container h1 {
  color: #e9edef;
  font-size: 2rem;
  font-weight: 500;
  margin-bottom: 32px;
}

form {
  width: 100%;
  max-width: 400px;
  background-color: #202c33;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.form-group {
  margin-bottom: 24px;
  text-align: left;
}

.form-group label {
  display: block;
  color: #aebac1;
  font-size: 0.95rem;
  font-weight: 500;
  margin-bottom: 10px;
}

input {
  display: block;
  padding: 14px 16px;
  margin: 0 auto;
  width: 100%;
  border: 1px solid #2a3942;
  border-radius: 8px;
  background-color: #2a3942;
  color: #e9edef;
  font-size: 1rem;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

input::placeholder {
  color: #667781;
}

input:focus {
  outline: none;
  background-color: #1f2c33;
  border-color: #00a884;
}

button {
  width: 100%;
  padding: 14px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  cursor: pointer;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s ease;
  margin-top: 8px;
}

button:hover {
  background-color: #06cf9c;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

button:active {
  transform: translateY(0);
}

.error-message {
  color: #ea4335;
  background-color: rgba(234, 67, 53, 0.15);
  border: 1px solid #ea4335;
  border-radius: 8px;
  padding: 12px 20px;
  margin-top: 20px;
  font-weight: 500;
  max-width: 400px;
  width: 100%;
  text-align: center;
}
</style>