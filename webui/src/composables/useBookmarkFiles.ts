import { computed, type Ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
    listBookmarkFiles,
    getBookmarkFile,
    deleteBookmarkFile,
    saveBookmarkFile,
    type BookmarkFileItem,
} from '@/lib/api/bookmarkFiles'
import { invalidateAndRefetch } from '@/composables/queryUtils'

// Every bookmark-file query starts with this key, so a save or delete
// refreshes the list and any widget preview showing the file.
const BOOKMARK_FILES_KEY = ['data', 'bookmarks']

/** The list of bookmark files, plus save/delete for the admin page. */
export function useBookmarkFiles() {
    const queryClient = useQueryClient()
    const doInvalidate = () => invalidateAndRefetch(queryClient, BOOKMARK_FILES_KEY)

    const query = useQuery({
        queryKey: [...BOOKMARK_FILES_KEY, 'list'],
        queryFn: listBookmarkFiles,
    })

    const saveMutation = useMutation({
        mutationFn: ({ name, bookmarks }: { name: string; bookmarks: BookmarkFileItem[] }) => saveBookmarkFile(name, bookmarks),
        onSuccess: doInvalidate,
    })

    const deleteMutation = useMutation({
        mutationFn: (name: string) => deleteBookmarkFile(name),
        onSuccess: doInvalidate,
    })

    return {
        files: computed(() => query.data.value ?? []),
        isLoading: query.isLoading,
        isError: query.isError,
        refetch: doInvalidate,

        saveFile: saveMutation.mutateAsync,
        deleteFile: deleteMutation.mutateAsync,
    }
}

/** The parsed bookmarks of one file, for a widget preview. Disabled while name is empty. */
export function useBookmarkFile(name: Ref<string>) {
    const query = useQuery({
        queryKey: computed(() => [...BOOKMARK_FILES_KEY, 'file', name.value]),
        queryFn: () => getBookmarkFile(name.value),
        enabled: computed(() => name.value !== ''),
        retry: false,
    })
    return {
        bookmarks: computed(() => query.data.value ?? []),
        isLoading: query.isLoading,
        isError: query.isError,
    }
}
