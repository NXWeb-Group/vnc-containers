<script setup>
import { ref } from 'vue'
import RFB from '@novnc/novnc/core/rfb.js';

const status = ref(null);

async function connect(path) {
  const url = (location.protocol === 'https:' ? 'wss' : 'ws') + '://' + location.host + `/websockify/${path}`;

  console.log(url);

  const rfb = new RFB(
    document.getElementById('screen'), url);

  rfb.addEventListener("connect", () => {
    status.value = "Connected";
  });
  rfb.addEventListener("disconnect", () => {
    status.value = "Disconnected";
  });
}

function getVNC() {
  status.value = "Starting container";
  fetch('/api/createContainer')
    .then(response => {
      if (response.ok) {
        return response.json();
      }
      throw new Error('Network response was not ok');
    })
    .then(data => {
      console.log(data);
      setTimeout(() => {
        connect(data.id);
      }, 5000);

    })
    .catch(error => {
      console.error('There was a problem with the fetch operation:', error);
    });
}

</script>

<template>
  <button v-if="!status" @click="getVNC">Get VNC</button>
  <div v-else class="top_bar">
    <div class="status">{{ status }}</div>
  </div>


  <div id="screen">
  </div>
</template>

<style scoped>

.status {
  text-align: center;
}

#screen {
  flex: 1;
  overflow: hidden;
}

.top_bar {
  background-color: #6e84a3;
  color: white;
  font: bold 12px Helvetica;
  padding: 6px 5px 4px 5px;
  border-bottom: 1px outset;
}
</style>