from jinja2 import Template
from subprocess import call

unique_items = """
func UniqueItems_{{ typeA.capitalize() }}Slice(items []{{ typeA }}) []{{ typeA }} {
    if len(items) == 0 {
        return items
    }
    visited := make(map[{{ typeA }}]struct{})
    uItems := make([]{{ typeA }}, 0, len(items))
    for _, item := range items {
        if _, ok := visited[item]; !ok {
            uItems = append(uItems, item)
            visited[item] = struct{}{}
        }
    }
    return uItems
}
"""

eq_items = """
func Equal_{{ typeA.capitalize() }}Slice(a, b []{{ typeA }}) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}
"""

eq_maps = """
func Equal_{{ typeA.capitalize() }}{{ typeB.capitalize() }}Map(a, b map[{{typeA}}]{{typeB}}) bool {
    if len(a) != len(b) {
        return false
    }
    for k := range a {
        if a[k] != b[k] {
            return false
        }
    }
    return true
}
"""

set_diff = """
func SetDiff_{{ typeA.capitalize() }}{{ typeB.capitalize() }}Map(a, b map[{{ typeA }}]{{ typeB }}) map[{{ typeA }}]{{ typeB }} {
    diff := make(map[{{ typeA }}]{{ typeB }})
    for k, v := range a {
        if _, ok := b[k]; !ok {
            diff[k] = v
        }
    }
    return diff
}
"""

clone_items = """
func Clone_{{ typeA.capitalize() }}Slice(a []{{ typeA }}) []{{ typeA }} {
    b := make([]{{ typeA }}, len(a))
    copy(b, a)
    return b
}
"""

clone_maps = """
func Clone_{{ typeA.capitalize() }}{{ typeB.capitalize() }}Map(a map[{{ typeA}}]{{ typeB }}) map[{{ typeA }}]{{ typeB }} {
    b := make(map[{{ typeA }}]{{ typeB }})
    for k, v := range a {
        b[k] = v
    }
    return b
}
"""

code_fragments = [
    {
        "template": Template(unique_items),
        "types": ["int64", "string", "int", "float64"],
    },
    {
        "template": Template(eq_items),
        "types": ["int64", "string", "int", "float64"],
    },
    {
        "template": Template(eq_maps),
        "types": [("string", "string")],
    },
    {
        "template": Template(set_diff),
        "types": [("string", "string")],
    },
    {
        "template": Template(clone_items),
        "types": ["int64", "string", "int", "float64"],
    },
    {
        "template": Template(clone_maps),
        "types": [("string", "string")],
    },
]

if __name__ == "__main__":
    type_names = ["type" + chr(x) for x in range(ord('A'), ord('Z')+1)]
    out = open("gen_code.go", "w")
    out.write("package generics\n")

    for fragment in code_fragments:
        for gt in fragment["types"]:
            ctx = {}
            if isinstance(gt, (list, tuple)):
                type_mapping = zip(type_names, gt)
                for tm in type_mapping:
                    ctx[tm[0]] = tm[1]
            else:
                ctx["typeA"] = gt
            out.write(fragment["template"].render(**ctx))
            out.write("\n")

    out.close()

    call(["goimports", "-e", "-w", "gen_code.go"])
