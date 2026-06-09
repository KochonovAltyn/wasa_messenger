<template>
  <div class="home-container">
    <h1>Home Page</h1>
    <p>Welcome, {{ username }}!</p>
    <!-- ✅ Display Profile Photo if Available -->
    <img v-if="profilePhoto" :src="profilePhoto" alt="Profile Photo" class="profile-photo" />
    <RouterLink to="/users/me/username" class="nav-button">Profile</RouterLink>
    <RouterLink to="/users/me/photo" class="photo-button">Upload Profile Photo</RouterLink>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "HomeView",
  data() {
    return {
      username: localStorage.getItem("username") || "Guest",
      profilePhoto: "",
    };
  },
  async created() {
    await this.fetchProfilePhoto();
  },
  methods: {
    async fetchProfilePhoto() {
      const token = localStorage.getItem("authToken");
      const userID = localStorage.getItem("userID");
      if (!userID || !token) { 
        console.warn("🚨 Missing user ID or token."); 
        return; 
      }
      try {
        console.log("🔍 Fetching profile photo from API...");
        const response = await axios.get(`/users/${userID}`, { 
          headers: { Authorization: `Bearer ${token}` } 
        });
        console.log("📝 Full API Response:", response.data);
        
        if (response.status === 404) { 
          console.error("🚨 User not found in database."); 
          return; 
        }
        
        if (response.data.photo) {
          console.log("📸 Raw photo value from API:", response.data.photo);
          
          if (typeof response.data.photo === "string") {
            this.profilePhoto = axios.defaults.baseURL + response.data.photo;
            localStorage.setItem(`profilePhoto_${userID}`, this.profilePhoto);
            console.log("✅ Final profile photo URL:", this.profilePhoto);
          } else {
            console.error("❌ API photo format is incorrect (expected a string but got an object)");
            console.log("🧐 Actual type:", typeof response.data.photo, "| Value:", response.data.photo);
            this.profilePhoto = "";
          }
        } else { 
          console.warn("🚨 No profile photo found for user."); 
          this.profilePhoto = ""; 
        }
      } catch (error) { 
        console.error("❌ Error fetching profile photo:", error); 
        this.profilePhoto = ""; 
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.home-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 60px 20px 40px;
  color: #e9edef;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.home-container h1 {
  color: #e9edef;
  font-size: 2rem;
  font-weight: 500;
  margin-bottom: 16px;
}

.home-container p {
  color: #aebac1;
  font-size: 1.1rem;
  margin-bottom: 32px;
}

.profile-photo {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  object-fit: cover;
  border: 4px solid #00a884;
  margin-bottom: 32px;
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.3);
}

.nav-button, 
.photo-button {
  display: block;
  margin: 12px auto;
  padding: 14px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  text-align: center;
  text-decoration: none;
  width: 240px;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s ease;
}

.nav-button:hover, 
.photo-button:hover {
  background-color: #06cf9c;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

.nav-button:active, 
.photo-button:active {
  transform: translateY(0);
}
</style>