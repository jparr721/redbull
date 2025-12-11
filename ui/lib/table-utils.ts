import { ColumnDef } from "@tanstack/react-table"
import { z } from "zod"

export function createColumnsFromSchema<T extends z.ZodObject<z.ZodRawShape>>(
  schema: T,
  options?: { exclude?: string[] }
): ColumnDef<z.infer<T>>[] {
  return Object.keys(schema.shape)
    .filter((key) => !options?.exclude?.includes(key))
    .map((key) => ({
      accessorKey: key,
      header: key.charAt(0).toUpperCase() + key.slice(1).replace(/([A-Z])/g, " $1"),
    }))
}
