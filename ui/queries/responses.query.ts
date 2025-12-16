import axios from 'axios';
import * as z from 'zod';

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
  const { data } = await axios.get<Response[]>("http://localhost:8000/responses");
  return data;
}
