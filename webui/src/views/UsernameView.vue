<template>
  <div class="username-container">
    <h1>Update Your Username</h1>
    <form @submit.prevent="updateUsername">
      <div class="form-group">
        <label for="newname">New Username</label>
        <input
          type="text"
          id="newname"
          v-model="newUsername"
          placeholder="Enter your new username"
          required
        />
      </div>
      <!-- ✅ Button will only be disabled if user is NOT logged in OR the input is empty -->
      <button type="submit" :disabled="!isLoggedIn || newUsername.trim() === ''">Update</button>
    </form>
    <div v-if="!isLoggedIn" class="error-message">
      You must be logged in to update your username.
    </div>
    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "UsernameView",
  data() {
    return { 
      newUsername: "", 
      isLoggedIn: false, 
      successMessage: "", 
      errorMessage: "" 
    };
  },
  created() {
    const token = localStorage.getItem("authToken");
    if (token) { 
      this.isLoggedIn = true; 
    } else { 
      this.isLoggedIn = false; 
      this.$router.push("/"); 
    }
  },
  methods: {
    async updateUsername() {
      this.errorMessage = "";
      const token = localStorage.getItem("authToken");
      if (!token) { 
        this.errorMessage = "You must be logged in."; 
        return; 
      }
      try {
        console.log("Updating username:", this.newUsername);
        console.log("Auth Token:", token);
        const response = await axios.put(`/users/me/username?t=${new Date().getTime()}`, 
          { newname: this.newUsername }, 
          {
            headers: { 
              "Authorization": `Bearer ${token}`, 
              "Content-Type": "application/json" 
            }
          }
        );
        console.log("Response:", response);
        localStorage.setItem("username", this.newUsername);
        this.successMessage = "Username updated successfully!";
        this.errorMessage = "";
        setTimeout(() => { 
          this.$router.push("/home"); 
        }, 1000);
      } catch (error) {
        console.error("Error updating username:", error);
        this.errorMessage = error.response?.data?.error || "An error occurred. Please try again.";
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.username-container {
  max-width: 450px;
  margin: 60px auto;
  padding: 32px 24px;
  text-align: center;
  border: 1px solid #2a3942;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  background-color: #202c33;
  color: #e9edef;
}

.username-container h1 {
  color: #e9edef;
  font-size: 1.75rem;
  font-weight: 500;
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 24px;
  text-align: left;
}

label {
  display: block;
  margin-bottom: 10px;
  font-weight: 500;
  color: #aebac1;
  font-size: 0.95rem;
}

input {
  width: 100%;
  padding: 12px 16px;
  font-size: 1rem;
  border: 1px solid #2a3942;
  border-radius: 8px;
  background-color: #2a3942;
  color: #e9edef;
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
  font-size: 1rem;
  color: #111b21;
  background-color: #00a884;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
}

button:hover:not(:disabled) {
  background-color: #06cf9c;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

button:active:not(:disabled) {
  transform: translateY(0);
}

button:disabled {
  background-color: #374955;
  color: #667781;
  cursor: not-allowed;
  transform: none;
}

.success-message {
  margin-top: 20px;
  padding: 12px 20px;
  background-color: rgba(0, 168, 132, 0.2);
  color: #00a884;
  border-left: 4px solid #00a884;
  border-radius: 8px;
  font-weight: 500;
}

.error-message {
  margin-top: 20px;
  padding: 12px 20px;
  background-color: rgba(234, 67, 53, 0.2);
  color: #ea4335;
  border-left: 4px solid #ea4335;
  border-radius: 8px;
  font-weight: 500;
}
</style>

