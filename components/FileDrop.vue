<script setup lang="ts">

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ref } from 'vue'

const emit = defineEmits(['file-selected'])

const isDragging = ref(false)

const handleFileInput = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files.length > 0) {
    emit('file-selected', input.files[0]);
  }
};

const handleDragOver = (event: DragEvent) => {
  event.preventDefault();
  isDragging.value = true;
};

const handleDragLeave = (event: DragEvent) => {
  event.preventDefault();
  isDragging.value = false;
};

const handleDrop = (event: DragEvent) => {
  event.preventDefault();
  isDragging.value = false;
  
  if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
    emit('file-selected', event.dataTransfer.files[0]);
  }
};
</script>

<template>
  <div>
    <div 
      class="border-2 border-dashed rounded-xl p-6 text-center transition-colors bg-white"
      :class="[
        isDragging ? 'border-sky-700 bg-sky-50' : 'border-gray-300 hover:border-sky-400',
        'cursor-pointer'
      ]"
      @dragover="handleDragOver"
      @dragleave="handleDragLeave"
      @drop="handleDrop"
    >
      <div class="space-y-4">
        <div class="text-gray-900">
          <p class="md:text-lg lg:text-xl font-medium">Drag and drop your GoPro file here</p>
        </div>
        <p class="space-y-4 text-md">or</p>
        <div class="flex flex-row justify-center items-center gap-4">
          <Label for="videofile" class="cursor-pointer bg-sky-700 hover:bg-sky-800 text-white rounded px-4 py-2">Select a file</Label>
          <Input 
            id="videofile" 
            type="file" 
            @change="handleFileInput" 
            accept="video/mp4" 
            class="hidden" 
          />
        </div>
      </div>
    </div>
  </div>
</template>

