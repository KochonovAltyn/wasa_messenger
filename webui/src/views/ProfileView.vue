<template>
  <div class="profile-container">
    <h1>Profile</h1>
    <div v-if="isLoggedIn">
      <!-- Display Profile Picture using the computed fullProfilePhoto -->
      <img 
        v-if="profilePhoto" 
        :src="fullProfilePhoto" 
        alt="Profile Photo" 
        class="profile-photo" 
      />
      
      <!-- Display Username -->
      <div class="username-section">
        <h2 class="current-username">@{{ username }}</h2>
        <button @click="showUsernameEdit = !showUsernameEdit" class="edit-username-btn">
          {{ showUsernameEdit ? 'Cancel' : 'Change Username' }}
        </button>
      </div>

      <!-- Username Edit Form -->
      <div v-if="showUsernameEdit" class="username-edit-form">
        <input 
          type="text" 
          v-model="newUsername" 
          placeholder="Enter new username"
          class="username-input"
        />
        <button @click="updateUsername" :disabled="!newUsername.trim()" class="save-username-btn">
          Save Username
        </button>
      </div>

      <!-- Photo Upload Section -->
      <div class="photo-upload-section">
        <h3>Update Profile Photo</h3>
        <input type="file" @change="handleFileUpload" accept="image/*" />
        <button @click="updatePhoto" :disabled="!selectedFile" class="update-button">
          Update Photo
        </button>
      </div>

      <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
      <p v-if="successMessage" class="success-message">{{ successMessage }}</p>
    </div>
    <div v-else class="error-message">
      User not logged in
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "ProfileView",
  data() {
    return {
      selectedFile: null,
      profilePhoto: "",
      username: "",
      newUsername: "",
      showUsernameEdit: false,
      isLoggedIn: false,
      errorMessage: "",
      successMessage: "",
    };
  },
  computed: {
    // Computes the full URL for the profile photo using axios.defaults.baseURL
    fullProfilePhoto() {
      const baseURL = axios.defaults.baseURL;
      return `${baseURL}${this.profilePhoto}`;
    }
  },
  created() {
    const token = localStorage.getItem("authToken");
    if (token) {
      this.isLoggedIn = true;
      this.fetchUserProfile();
    } else {
      this.$router.push("/");
    }
  },
  methods: {
    handleFileUpload(event) {
      this.selectedFile = event.target.files[0];
    },

    async fetchUserProfile() {
      const token = localStorage.getItem("authToken");
      try {
        const response = await axios.get("/users/me", {
          headers: { Authorization: `Bearer ${token}` }
        });
        
        if (response.data.photo) {
          this.profilePhoto = response.data.photo;
          localStorage.setItem("profilePhoto", response.data.photo);
        }
        
        if (response.data.username) {
          this.username = response.data.username;
          localStorage.setItem("username", response.data.username);
        }
      } catch (error) {
        console.error("Error fetching user profile:", error);
        this.errorMessage = "Failed to load profile information";
      }
    },

    async updateUsername() {
      if (!this.newUsername.trim()) {
        this.errorMessage = "Username cannot be empty";
        return;
      }

      const token = localStorage.getItem("authToken");
      
      try {
        this.errorMessage = "";
        this.successMessage = "";
        
        await axios.put("/users/me/username", 
          { username: this.newUsername.trim() },
          { headers: { Authorization: `Bearer ${token}` } }
        );
        
        this.username = this.newUsername.trim();
        localStorage.setItem("username", this.username);
        this.successMessage = "Username updated successfully!";
        this.showUsernameEdit = false;
        this.newUsername = "";
        
      } catch (error) {
        // Improved error handling
        if (error.response?.status === 409) {
          this.errorMessage = "This username is already taken";
        } else if (error.response?.status === 400) {
          this.errorMessage = "Invalid username format";
        } else {
          this.errorMessage = error.response?.data?.error || "An error occurred. Please try again.";
        }
      }
    },

    async updatePhoto() {
      if (!this.selectedFile) {
        this.errorMessage = "Please select an image first.";
        return;
      }
      
      const token = localStorage.getItem("authToken");
      const userID = localStorage.getItem("userID");
      
      if (!token || !userID) {
        this.errorMessage = "You must be logged in.";
        return;
      }
      
      const formData = new FormData();
      formData.append("photo", this.selectedFile);
      
      try {
        this.errorMessage = "";
        this.successMessage = "";
        
        const response = await axios.put("/users/me/photo", formData, {
          headers: { 
            Authorization: `Bearer ${token}`, 
            "Content-Type": "multipart/form-data" 
          }
        });
        
        if (response.data.photo) {
          this.profilePhoto = response.data.photo;
          localStorage.setItem(`profilePhoto_${userID}`, response.data.photo);
          this.successMessage = "Profile photo updated successfully!";
          this.selectedFile = null;
          
          // Redirect to home after 1.5 seconds
          setTimeout(() => {
            this.$router.push("/home");
          }, 1500);
        }
      } catch (error) {
        this.errorMessage = error.response?.data?.error || "An error occurred while updating photo.";
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.profile-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 60px 20px 40px;
  color: #e9edef;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.profile-container h1 {
  color: #e9edef;
  font-size: 2rem;
  font-weight: 500;
  margin-bottom: 32px;
}

.profile-photo {
  width: 180px;
  height: 180px;
  border-radius: 50%;
  object-fit: cover;
  border: 4px solid #00a884;
  margin-bottom: 24px;
  box-shadow: 0 4px 16px rgba(0, 168, 132, 0.3);
}

.username-section {
  text-align: center;
  margin-bottom: 32px;
}

.current-username {
  color: #00a884;
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 16px;
}

.edit-username-btn {
  padding: 10px 24px;
  background-color: #202c33;
  color: #00a884;
  border: 1px solid #00a884;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.edit-username-btn:hover {
  background-color: #2a3942;
}

.username-edit-form {
  background-color: #202c33;
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 32px;
  max-width: 400px;
  width: 100%;
}

.username-input {
  width: 100%;
  padding: 12px 16px;
  background-color: #2a3942;
  border: 1px solid #2a3942;
  border-radius: 8px;
  color: #e9edef;
  font-size: 0.95rem;
  margin-bottom: 12px;
  box-sizing: border-box;
}

.username-input:focus {
  outline: none;
  border-color: #00a884;
}

.username-input::placeholder {
  color: #667781;
}

.save-username-btn {
  width: 100%;
  padding: 12px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
}

.save-username-btn:hover:not(:disabled) {
  background-color: #06cf9c;
}

.save-username-btn:disabled {
  background-color: #374955;
  color: #667781;
  cursor: not-allowed;
}

.photo-upload-section {
  background-color: #202c33;
  padding: 24px;
  border-radius: 12px;
  max-width: 400px;
  width: 100%;
}

.photo-upload-section h3 {
  color: #e9edef;
  font-size: 1.2rem;
  margin-bottom: 16px;
  text-align: center;
}

input[type="file"] {
  display: block;
  margin: 20px auto;
  padding: 12px 16px;
  background-color: #2a3942;
  border: 1px solid #2a3942;
  border-radius: 8px;
  color: #e9edef;
  cursor: pointer;
  width: 100%;
  font-size: 0.95rem;
  box-sizing: border-box;
}

input[type="file"]::file-selector-button {
  background-color: #00a884;
  color: #111b21;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  margin-right: 12px;
  transition: background-color 0.2s ease;
}

input[type="file"]::file-selector-button:hover {
  background-color: #06cf9c;
}

.update-button {
  width: 100%;
  margin-top: 12px;
  padding: 14px 32px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s ease;
}

.update-button:hover:not(:disabled) {
  background-color: #06cf9c;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

.update-button:active:not(:disabled) {
  transform: translateY(0);
}

.update-button:disabled {
  background-color: #374955;
  color: #667781;
  cursor: not-allowed;
  transform: none;
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

.success-message {
  color: #00a884;
  background-color: rgba(0, 168, 132, 0.15);
  border: 1px solid #00a884;
  border-radius: 8px;
  padding: 12px 20px;
  margin-top: 20px;
  font-weight: 500;
  max-width: 400px;
  width: 100%;
  text-align: center;
}
</style>

