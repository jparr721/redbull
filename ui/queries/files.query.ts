import axios from "axios";
import * as z from 'zod';
import { API_BASE_URL } from "@/lib/api-config";

export const FileInfoSchema = z.object({
    name: z.string(),
    size: z.number(),
    modTime: z.string(),
});

export type FileInfo = z.infer<typeof FileInfoSchema>;

export async function getFiles(): Promise<FileInfo[]> {
    const { data } = await axios.get<FileInfo[]>(`${API_BASE_URL}/files`);
    return data;
}