<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useStore, type GpsData } from "~/store";
import { Input } from '@/components/ui/input'

const store = useStore()

const fileInput = ref(null);

const handleFile = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (!input.files || input.files.length === 0) {
    console.error("No file selected");
    return;
  }

  const file = input.files[0];

  store.setVideoUrl(URL.createObjectURL(file));

  if (file.type !== "video/mp4" && !file.name.endsWith(".mp4")) {
    console.error("Selected file is not an MP4");
    return;
  }

  if (window.processFile) {
    try {
      const gpsData = await window.processFile(file) as GpsData[];
      store.setGpsData(gpsData);
    } catch (err) {
      console.error("Error processing file:", err);
    }
  } else {
    console.error("WASM processFile function not available");
  }
};


</script>

<template>
  <div>
    <h1>GoPro File Upload</h1>
    <Input type="file" @change="handleFile" accept="video/mp4" />
  </div>
</template>

