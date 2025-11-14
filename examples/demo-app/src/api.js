// API functions for the demo app
import { formatDate } from './utils.js';

export async function fetchData(endpoint) {
  try {
    const response = await fetch(endpoint);
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching data:', error);
    throw error;
  }
}

export function processData(data) {
  return data.map(item => ({
    ...item,
    formattedDate: formatDate(item.date)
  }));
}
