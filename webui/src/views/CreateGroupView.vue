<template>
  <div class="create-group-container">
    <h1>Create a New Group</h1>
    <form @submit.prevent="createGroup">
      <div class="form-group">
        <label for="groupName">Group Name</label>
        <input type="text" id="groupName" v-model="groupName" placeholder="Enter group name" required />
      </div>
      <div class="form-group">
        <label for="groupPhoto">Group Photo (Optional)</label>
        <input type="file" id="groupPhoto" @change="handlePhotoChange" />
      </div>
      <div class="form-group">
        <label for="usernames">Add Members</label>
        <input type="text" id="usernames" v-model="usernamesInput" placeholder="Enter usernames (comma separated)" required />
      </div>
      <button type="submit" :disabled="loading">
        {{ loading ? 'Creating...' : 'Create Group' }}
      </button>
    </form>
    <div v-if="message" :class="messageClass">
      {{ message }}
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "CreateGroupView",
  data() {
    return {
      groupName: '',
      groupPhoto: null,
      usernamesInput: '',
      loading: false,
      message: '',
      messageClass: ''
    };
  },
  methods: {
    handlePhotoChange(event) {
      this.groupPhoto = event.target.files[0];
    },
    async createGroup() {
      this.loading = true;
      this.message = '';
      this.messageClass = '';
      
      const token = localStorage.getItem('authToken');
      const userID = localStorage.getItem('userID');
      
      if (!token || !userID) {
        this.message = 'User not authenticated';
        this.messageClass = 'error';
        this.loading = false;
        return;
      }

      const usernames = this.usernamesInput
        .split(',')
        .map(username => username.trim())
        .filter(username => username.length > 0); // Remove empty strings

      if (usernames.length === 0) {
        this.message = 'Please add at least one member';
        this.messageClass = 'error';
        this.loading = false;
        return;
      }

      const formData = new FormData();
      formData.append('group_name', this.groupName);
      if (this.groupPhoto) {
        formData.append('photo', this.groupPhoto);
      }
      formData.append('usernames', JSON.stringify(usernames));

      try {
        const response = await axios.post('/groups', formData, {
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'multipart/form-data',
          },
        });

        this.loading = false;
        const groupId = response.data.c_id;
        
        this.message = `Group "${this.groupName}" created successfully!`;
        this.messageClass = 'success';

        // ✅ REDIRECT TO THE NEW GROUP CONVERSATION
        setTimeout(() => {
          this.$router.push(`/chat/${groupId}`);
        }, 1000);

      } catch (error) {
        this.loading = false;
        
        // Better error handling
        if (error.response?.status === 409) {
          this.message = 'A group with this name already exists';
        } else if (error.response?.status === 404) {
          this.message = 'One or more usernames not found';
        } else {
          this.message = error.response?.data?.error || 'Failed to create group. Please try again.';
        }
        
        this.messageClass = 'error';
        console.error("Error creating group:", error);
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.create-group-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 40px 20px;
  color: #e9edef;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.create-group-container h1 {
  color: #e9edef;
  font-size: 1.75rem;
  font-weight: 500;
  margin-bottom: 32px;
  text-align: center;
}

form {
  width: 100%;
  max-width: 500px;
  background-color: #202c33;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.form-group {
  margin-bottom: 20px;
  text-align: left;
}

.form-group label {
  display: block;
  color: #aebac1;
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 8px;
}

input[type="text"],
input[type="file"] {
  width: 100%;
  padding: 12px 16px;
  margin: 0;
  background-color: #2a3942;
  border: 1px solid #2a3942;
  border-radius: 8px;
  color: #e9edef;
  font-size: 0.95rem;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

input[type="text"]:focus {
  outline: none;
  background-color: #1f2c33;
  border-color: #00a884;
}

input[type="text"]::placeholder {
  color: #667781;
}

input[type="file"] {
  padding: 10px 16px;
  cursor: pointer;
}

input[type="file"]::file-selector-button {
  background-color: #00a884;
  color: #111b21;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  margin-right: 12px;
  transition: background-color 0.2s ease;
}

input[type="file"]::file-selector-button:hover {
  background-color: #06cf9c;
}

button {
  width: 100%;
  padding: 14px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s ease;
  margin-top: 8px;
}

button:hover:not(:disabled) {
  background-color: #06cf9c;
  transform: translateY(-1px);
}

button:disabled {
  background-color: #374955;
  color: #667781;
  cursor: not-allowed;
  transform: none;
}

.message {
  margin-top: 20px;
  padding: 12px 20px;
  border-radius: 8px;
  font-weight: 500;
  text-align: center;
  max-width: 500px;
  width: 100%;
}

.message.success {
  background-color: rgba(0, 168, 132, 0.2);
  color: #00a884;
  border-left: 4px solid #00a884;
}

.message.error {
  background-color: rgba(234, 67, 53, 0.2);
  color: #ea4335;
  border-left: 4px solid #ea4335;
}
</style>

