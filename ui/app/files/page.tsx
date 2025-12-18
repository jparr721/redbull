"use client";

import { useState } from 'react';
import { FileInfo, getFiles } from "@/queries/files.query";
import { useQuery } from "@tanstack/react-query";
import { FilesTable } from './_components/FilesTable';

export default function Files() {
    const [search, setSearch] = useState("");
    const { data: files, isLoading } = useQuery({
        queryKey: ["files"],
        queryFn: getFiles,
    });

    if (isLoading) return <h1>Loading</h1>;

    return (
    <div>
        <FilesTable files={files} />
    </div>
    );
}