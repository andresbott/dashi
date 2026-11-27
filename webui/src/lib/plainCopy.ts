// plainCopy returns a detached deep copy of JSON data.
//
// A JSON round-trip, not structuredClone: editor state can hold reactive
// proxies below the top level (rows and configs are rebuilt around existing
// children), which toRaw does not reach and structuredClone rejects. The data
// copied here is plain JSON anyway: it is stored and sent as JSON.
export function plainCopy<T>(value: T): T {
    return JSON.parse(JSON.stringify(value)) as T
}
