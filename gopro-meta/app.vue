<script setup lang="ts">
import { onMounted } from "vue";

let go;
let wasmInstance;

const loadWasmExec = async () => {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = "/wasm_exec.js"; // Load from public folder
    script.onload = resolve;
    script.onerror = reject;
    document.body.appendChild(script);
  });
};

// Load wasm_exec.js and WebAssembly
onMounted(async () => {
  try {
    await loadWasmExec();
    go = new Go();
    const wasmResponse = await fetch("/main.wasm");
    const wasmBytes = await wasmResponse.arrayBuffer();
    const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);
    go.run(instance);
    wasmInstance = instance;
  } catch (error) {
    console.error("Error loading WebAssembly:", error);
  }
});
</script>

<template>
  <section class="mb-8 p-4">
    <h1>TrackBack</h1>
  </section>
  <section class="mx-4 flex flex-row  gap-x-4">
    <div>
      <GoProUpload class="mb-4" />
      <Video />
    </div>
    <div class="h-[600px] flex-1">
      <Map />
    <AccelerationVisualizer />
    </div>
  </section>
</template>
