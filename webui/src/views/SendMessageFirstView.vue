<template>
  <div class="send-message-container">
    <h1>Start a New Conversation</h1>

    <!-- Loading state -->
    <div v-if="loading" class="loading-message">Sending message...</div>

    <!-- Form for text message and file upload -->
    <div v-else>
      <form @submit.prevent="submitForm">
        <div class="form-group">
          <label for="recipient_username">Recipient Username</label>
          <input 
            v-model="recipientUsername" 
            type="text" 
            id="recipient_username" 
            required
            placeholder="Enter recipient's username"
          />
        </div>

        <!-- Choose message type: text, image, or gif -->
        <div class="form-group">
          <label for="message_type">Message Type</label>
          <select v-model="messageType" id="message_type" required>
            <option disabled value="">Select message type</option>
            <option value="text">Text</option>
            <option value="image">Image</option>
            <option value="gif">GIF</option>
          </select>
        </div>

        <!-- Message content section (only shown if the user selects 'text') -->
        <div v-if="messageType === 'text'" class="form-group">
          <label for="content">Message</label>
          <textarea 
            v-model="content" 
            id="content" 
            required
            placeholder="Type your message here..."
          ></textarea>
        </div>

        <!-- File upload section (only shown if the user selects 'image' or 'gif') -->
        <div v-if="messageType === 'image' || messageType === 'gif'" class="form-group">
          <label for="file">Upload a file (Image/GIF)</label>
          <input 
            type="file" 
            id="file" 
            accept="image/*, .gif"
            @change="handleFileChange"
            required
          />
        </div>

        <div class="form-group">
          <button type="submit" class="submit-btn" :disabled="loading">
            {{ loading ? 'Sending...' : 'Send Message' }}
          </button>
        </div>
      </form>
    </div>

    <!-- Success or Error Message -->
    <div v-if="message">
      <p :class="messageClass">{{ message }}</p>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  data() {
    return {
      recipientUsername: "",
      content: "",
      file: null,
      messageType: "",
      loading: false,
      message: "",
      messageClass: ""
    };
  },
  methods: {
    handleFileChange(event) { 
      this.file = event.target.files[0]; 
    },
    async submitForm() {
      const token = localStorage.getItem("authToken");
      const userID = localStorage.getItem("userID");
      
      if (!token || !userID) { 
        this.message = "User not authenticated."; 
        this.messageClass = "error"; 
        return; 
      }

      const formData = new FormData();
      formData.append("recipient_username", this.recipientUsername);
      formData.append("content_type", this.messageType);
      formData.append("content", this.content);
      
      if (this.file) { 
        formData.append("file", this.file); 
      }

      try {
        this.loading = true; 
        this.message = ""; 
        this.messageClass = "";

        const response = await axios.post(
          `/users/${userID}/conversations/first-message`, 
          formData, 
          {
            headers: { 
              Authorization: `Bearer ${token}`, 
              "Content-Type": "multipart/form-data" 
            }
          }
        );

        this.loading = false;
        const conversationId = response.data.c_id;
        
        this.message = `Message sent successfully!`;
        this.messageClass = "success";

        // ✅ REDIRECT TO THE NEW CONVERSATION
        setTimeout(() => {
          this.$router.push(`/chat/${conversationId}`);
        }, 1000);

      } catch (error) {
        this.loading = false;
        
        // Better error handling
        if (error.response?.status === 404) {
          this.message = `User "${this.recipientUsername}" not found`;
        } else if (error.response?.status === 400) {
          this.message = "Invalid message format";
        } else {
          this.message = error.response?.data?.error || "Failed to send message. Please try again.";
        }
        
        this.messageClass = "error";
        console.error("Error sending message:", error);
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.send-message-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 40px 20px;
  color: #e9edef;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.send-message-container h1 {
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
  margin: 20px 0;
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
textarea,
select {
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
  font-family: inherit;
}

input[type="text"]::placeholder,
textarea::placeholder {
  color: #667781;
}

input[type="text"]:focus,
textarea:focus,
select:focus {
  outline: none;
  background-color: #1f2c33;
  border-color: #00a884;
}

textarea {
  min-height: 100px;
  resize: vertical;
}

select {
  cursor: pointer;
}

select option {
  background-color: #202c33;
  color: #e9edef;
}

input[type="file"] {
  width: 100%;
  padding: 10px 16px;
  background-color: #2a3942;
  border: 1px solid #2a3942;
  border-radius: 8px;
  color: #e9edef;
  cursor: pointer;
  font-size: 0.95rem;
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

button.submit-btn {
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

button.submit-btn:hover:not(:disabled) {
  background-color: #06cf9c;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

button.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

button.submit-btn:disabled {
  background-color: #374955;
  color: #667781;
  cursor: not-allowed;
  transform: none;
}

.loading-message {
  color: #aebac1;
  font-size: 1.1rem;
  margin-top: 20px;
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


