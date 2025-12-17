import axios from 'axios';
import * as z from 'zod';
import { API_BASE_URL } from '@/lib/api-config';

export const ResponseSchema = z.object({
  id: z.uuid(),
  time: z.string(),
  stdout: z.string(),
  stderr: z.string(),
  command: z.string(),
  currentDirectory: z.string(),
});

export type Response = z.infer<typeof ResponseSchema>;

export async function getResponses() {
  const { data } = await axios.get<Response[]>(`${API_BASE_URL}/responses`);
  return data;
}

export const LastCheckInSchema = z.object({
  checkInTime: z.string(),
});

export type LastCheckIn = z.infer<typeof LastCheckInSchema>;

export async function getLastCheckInTime() {
  const { data } = await axios.get<LastCheckIn>(`${API_BASE_URL}/last_checkin`);
  return data;
}