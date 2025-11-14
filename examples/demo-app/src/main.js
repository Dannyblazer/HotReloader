// Main application entry point
import { capitalize, debounce } from './utils.js';
import { fetchData, processData } from './api.js';

console.log('Demo App Starting...');

async function initialize() {
  console.log('Initializing application...');

  const appName = capitalize('hot reload optimizer demo');
  console.log(`App: ${appName}`);

  // Simulate some work
  const handleChange = debounce(() => {
    console.log('File changed, triggering rebuild...');
  }, 300);

  // This would normally fetch real data
  try {
    const data = await fetchData('https://api.example.com/data');
    const processed = processData(data);
    console.log('Data processed:', processed);
  } catch (error) {
    console.log('Running in demo mode (no API available)');
  }

  console.log(' Application initialized successfully');
}

initialize();
