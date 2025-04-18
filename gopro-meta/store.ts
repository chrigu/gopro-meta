import { defineStore } from 'pinia'


export interface GpsData {
  latitude: number;
  longitude: number;
  altitude: number;
  timestamp: number;
}

export interface AccelerationData {
  x: number;
  y: number;
  z: number;
  timestamp: number;
}

export interface FaceData {
  confidence: number;
  id: number;
  x: number;
  y: number;
  w: number;
  h: number;
  smile: number;
  blink: number;
  timestamp: number;
}

export interface LuminanceData {
  luminance: number;
  timestamp: number;
}

export interface HueData {
  hues: {
    hue: number;
    weight: number;
  }[];
  timestamp: number;
}

export interface ColorData {
  red: number;
  green: number;
  blue: number;
  timestamp: number;
}


export const useStore = defineStore('metaData', {
  state: () => ({
    gpsData: [] as GpsData[],
    accelerationData: [] as AccelerationData[],
    faceData: [] as FaceData[],
    luminanceData: [] as LuminanceData[],
    hueData: [] as HueData[],
    videoCurrentTime: 0,
    videoUrl: '',
  }),
  
  getters: {
    center(state): [number, number] {
      if (state.gpsData.length === 0) {
        return [47.21322, -1.559482]; // Default center
      }
      // Compute center as the average of all points
      const avgLat =
        state.gpsData.reduce((sum, p) => sum + p.latitude, 0) / state.gpsData.length;
      const avgLng =
        state.gpsData.reduce((sum, p) => sum + p.longitude, 0) / state.gpsData.length;
      return [avgLat, avgLng];
    },
    // todo: refactor
    currentGpsData(state) {
      const starTime = state.gpsData[0]?.timestamp;
      return findClosestObject(state.gpsData, state.videoCurrentTime, starTime);
    },
    currentAccelerationData(state) {
      const startTime = state.accelerationData[0]?.timestamp;
      return findClosestObject(state.accelerationData, state.videoCurrentTime, startTime);
    },
    currentFaceData(state) {
      const startTime = state.faceData[0]?.timestamp;
      return findClosestObject(state.faceData, state.videoCurrentTime, startTime);
    },
    currentLuminanceData(state) {
      const startTime = state.accelerationData[0]?.timestamp;
      return findClosestObject(state.luminanceData, state.videoCurrentTime, startTime);
    },
    currentHueData(state) {
      const startTime = state.hueData[0]?.timestamp;
      return findClosestObject(state.hueData, state.videoCurrentTime, startTime);
    },
  },

  actions: {
    setGpsData(data: GpsData[]) {
      this.gpsData = data;
    },
    setGyroData(data: AccelerationData[]) {
      this.accelerationData = data;
    },
    setFaceData(data: FaceData[]) {
      this.faceData = data.filter(face => face.confidence > 0 && face.x > 0 && face.y > 0 && face.w > 0 && face.h > 0);
    },
    setLuminanceData(data: LuminanceData[]) {
      this.luminanceData = data;
    },
    setHueData(data: HueData[]) {
      this.hueData = data.map(hue => ({
        hues: hue.hues.map((h: { hue: number; weight: number }) => ({
          hue: h.hue * 360 / 255,
          weight: h.weight * 100 / 255
        })),
        timestamp: hue.timestamp
      }))
    },
    setVideoCurrentTime(time: number) {
      this.videoCurrentTime = time;
    },
    setVideoUrl(url: string) {
      this.videoUrl = url;
    },
  },
});

const findClosestObject = (arr: { timestamp: number }[], targetTimestamp: number, starTime: number) => {
  if (arr.length === 0) return null;

  let left = 0;
  let right = arr.length - 1;

  while (left < right) {
    const mid = Math.floor((left + right) / 2);

    if ((arr[mid].timestamp - starTime)/1000 === targetTimestamp) {
      return arr[mid]; // Exact match
    } else if ((arr[mid].timestamp - starTime)/1000  < targetTimestamp) {
      left = mid + 1;
    } else {
      right = mid;
    }
  }

  // After the loop, 'left' is the closest index or the one just after the target
  if (left === 0) {
    const timeDiff = Math.abs((arr[0].timestamp - starTime)/1000 - targetTimestamp);
    return timeDiff <= 0.5 ? arr[0] : null;
  }
  if (left >= arr.length) {
    const timeDiff = Math.abs((arr[arr.length - 1].timestamp - starTime)/1000 - targetTimestamp);
    return timeDiff <= 0.5 ? arr[arr.length - 1] : null;
  }

  // Compare the two closest candidates
  const prev = arr[left - 1];
  const next = arr[left];
  
  const prevTimeDiff = Math.abs((prev.timestamp - starTime)/1000 - targetTimestamp);
  const nextTimeDiff = Math.abs((next.timestamp - starTime)/1000 - targetTimestamp);
  
  // Return the closer sample only if it's within 0.5s window
  if (prevTimeDiff <= 0.5 && nextTimeDiff <= 0.5) {
    return prevTimeDiff <= nextTimeDiff ? prev : next;
  } else if (prevTimeDiff <= 0.5) {
    return prev;
  } else if (nextTimeDiff <= 0.5) {
    return next;
  }
  
  return null; // No sample within ±0.5s window
};


function hueToRGB(h: number) {
  // H is in degrees: 0–360
  h = h % 360;
  const s = 1, l = 0.5;

  const c = (1 - Math.abs(2 * l - 1)) * s;
  const x = c * (1 - Math.abs((h / 60) % 2 - 1));
  const m = l - c/2;
  let r = 0, g = 0, b = 0;

  if (h < 60) [r, g, b] = [c, x, 0];
  else if (h < 120) [r, g, b] = [x, c, 0];
  else if (h < 180) [r, g, b] = [0, c, x];
  else if (h < 240) [r, g, b] = [0, x, c];
  else if (h < 300) [r, g, b] = [x, 0, c];
  else [r, g, b] = [c, 0, x];

  return [r + m, g + m, b + m];
}

function normalize(weights: number[]) {
  const total = weights.reduce((sum, w) => sum + w, 0);
  return weights.map(w => w / total);
}

function blendHues(huesRaw: number[], weights: number[]) {
  const normWeights = normalize(weights);
  const hues = huesRaw.map(h => h * 360 / 255);

  let r = 0, g = 0, b = 0;

  for (let i = 0; i < hues.length; i++) {
    const [ri, gi, bi] = hueToRGB(hues[i]);
    r += ri * normWeights[i];
    g += gi * normWeights[i];
    b += bi * normWeights[i];
  }

  return to255([r, g, b]);
}

function to255([r, g, b]: number[]) {
  return [Math.round(r * 255), Math.round(g * 255), Math.round(b * 255)];
}
