package compile

import (
	"v.io/lib/toposort"
	"v.io/v23/vdl"
	"v.io/v23/vdl/parse"
)

// compileInterfaces is the "entry point" to the rest of this file.  It takes
// the interfaces defined in pfiles and compiles them into Interfaces in pkg.
func compileInterfaces(pkg *Package, pfiles []*parse.File, env *Env) {
	id := ifaceDefiner{pkg, pfiles, env, make(map[string]*ifaceBuilder)}
	if id.Declare(); !env.Errors.IsEmpty() {
		return
	}
	id.SortAndDefine()
}

// ifaceDefiner defines interfaces in a package.  This is split into two phases:
// 1) Declare ensures local interface references can be resolved.
// 2) SortAndDefine sorts in dependency order, and evaluates and defines each
//    const.
//
// It holds a builders map from interface name to ifaceBuilder, where the
// ifaceBuilder is responsible for compiling and defining a single interface.
type ifaceDefiner struct {
	pkg      *Package
	pfiles   []*parse.File
	env      *Env
	builders map[string]*ifaceBuilder
}

type ifaceBuilder struct {
	def  *Interface
	pdef *parse.Interface
}

func printIfaceBuilderName(ibuilder interface{}) string {
	return ibuilder.(*ifaceBuilder).def.Name
}

// Declare creates builders for each interface defined in the package.
func (id ifaceDefiner) Declare() {
	for ix := range id.pkg.Files {
		file, pfile := id.pkg.Files[ix], id.pfiles[ix]
		for _, pdef := range pfile.Interfaces {
			export, err := ValidIdent(pdef.Name, ReservedNormal)
			if err != nil {
				id.env.prefixErrorf(file, pdef.Pos, err, "interface %s invalid name", pdef.Name)
				continue // keep going to catch more errors
			}
			detail := identDetail("interface", file, pdef.Pos)
			if err := file.DeclareIdent(pdef.Name, detail); err != nil {
				id.env.prefixErrorf(file, pdef.Pos, err, "interface %s name conflict", pdef.Name)
				continue
			}
			def := &Interface{NamePos: NamePos(pdef.NamePos), Exported: export, File: file}
			id.builders[pdef.Name] = &ifaceBuilder{def, pdef}
		}
	}
}

// Sort and define interfaces.  We sort by dependencies on other interfaces in
// this package.  The sorting is to ensure there are no cycles.
func (id ifaceDefiner) SortAndDefine() {
	// Populate sorter with dependency information.  The sorting ensures that the
	// list of interfaces within each file is topologically sorted, and also
	// deterministic; in the absence of interface embeddings, interfaces are
	// listed in the same order they were defined in the parsed files.
	var sorter toposort.Sorter
	for _, pfile := range id.pfiles {
		for _, pdef := range pfile.Interfaces {
			b := id.builders[pdef.Name]
			sorter.AddNode(b)
			for _, dep := range id.getLocalDeps(b) {
				sorter.AddEdge(b, dep)
			}
		}
	}
	// Sort and check for cycles.
	sorted, cycles := sorter.Sort()
	if len(cycles) > 0 {
		cycleStr := toposort.DumpCycles(cycles, printIfaceBuilderName)
		first := cycles[0][0].(*ifaceBuilder)
		id.env.Errorf(first.def.File, first.def.Pos, "package %v has cyclic interfaces: %v", id.pkg.Name, cycleStr)
		return
	}
	// Define all interfaces.  Since we add the interfaces as we go and evaluate
	// in topological order, dependencies are guaranteed to be resolvable when we
	// get around to defining the interfaces that embed on them.
	for _, ibuilder := range sorted {
		b := ibuilder.(*ifaceBuilder)
		id.define(b)
		addIfaceDef(b.def)
	}
}

// addIfaceDef updates our various structures to add a new interface.
func addIfaceDef(def *Interface) {
	def.File.Interfaces = append(def.File.Interfaces, def)
	def.File.Package.ifaceDefs[def.Name] = def
}

// getLocalDeps returns the list of interface dependencies for b that are in
// this package.
func (id ifaceDefiner) getLocalDeps(b *ifaceBuilder) (deps []*ifaceBuilder) {
	for _, pe := range b.pdef.Embeds {
		// Embeddings of other interfaces in this package are all we care about.
		if dep := id.builders[pe.Name]; dep != nil {
			deps = append(deps, dep)
		}
	}
	return
}

func (id ifaceDefiner) define(b *ifaceBuilder) {
	id.defineEmbeds(b)
	id.defineMethods(b)
}

func (id ifaceDefiner) defineEmbeds(b *ifaceBuilder) {
	// TODO(toddw): Check for duplicate methods.
	def, file := b.def, b.def.File
	seen := make(map[string]*parse.NamePos)
	for _, pe := range b.pdef.Embeds {
		if dup := seen[pe.Name]; dup != nil {
			id.env.Errorf(file, pe.Pos, "interface %s duplicate embedding (previous at %s)", pe.Name, dup.Pos)
			continue // keep going to catch more errors
		}
		seen[pe.Name] = pe
		// Resolve the embedded interface.
		embed, matched := id.env.ResolveInterface(pe.Name, file)
		if embed == nil {
			id.env.Errorf(file, pe.Pos, "interface %s undefined", pe.Name)
			continue // keep going to catch more errors
		}
		if len(matched) < len(pe.Name) {
			id.env.Errorf(file, pe.Pos, "interface %s invalid (%s unmatched)", pe.Name, pe.Name[len(matched):])
			continue // keep going to catch more errors
		}
		def.Embeds = append(def.Embeds, embed)
	}
}

func (id ifaceDefiner) defineMethods(b *ifaceBuilder) {
	def, file := b.def, b.def.File
	seen := make(map[string]*parse.Method)
	for _, pm := range b.pdef.Methods {
		if dup := seen[pm.Name]; dup != nil {
			id.env.Errorf(file, pm.Pos, "method %s redefined (previous at %s)", pm.Name, dup.Pos)
			continue // keep going to catch more errors
		}
		seen[pm.Name] = pm
		if err := ValidExportedIdent(pm.Name, ReservedCamelCase); err != nil {
			id.env.Errorf(file, pm.Pos, "method %s name (%s)", pm.Name, err)
			continue // keep going to catch more errors
		}
		m := &Method{NamePos: NamePos(pm.NamePos)}
		m.InArgs = id.defineArgs(in, m.NamePos, pm.InArgs, file)
		m.OutArgs = id.defineArgs(out, m.NamePos, pm.OutArgs, file)
		m.InStream = id.defineStreamType(pm.InStream, file)
		m.OutStream = id.defineStreamType(pm.OutStream, file)
		m.Tags = id.defineTags(pm.Tags, file)
		def.Methods = append(def.Methods, m)
	}
}

type inout string

const (
	in  inout = "in"
	out inout = "out"
)

func (id ifaceDefiner) defineArgs(io inout, method NamePos, pargs []*parse.Field, file *File) (args []*Field) {
	seen := make(map[string]*parse.Field)
	for _, parg := range pargs {
		if dup := seen[parg.Name]; dup != nil && parg.Name != "" {
			id.env.Errorf(file, parg.Pos, "method %s arg %s duplicate name (previous at %s)", method.Name, parg.Name, dup.Pos)
			continue // keep going to catch more errors
		}
		seen[parg.Name] = parg
		if io == out && len(pargs) > 2 && parg.Name == "" {
			id.env.Errorf(file, parg.Pos, "method %s out arg unnamed (must name all out args if there are more than 2)", method.Name)
			continue // keep going to catch more errors
		}
		if parg.Name != "" {
			if _, err := ValidIdent(parg.Name, ReservedCamelCase); err != nil {
				id.env.prefixErrorf(file, parg.Pos, err, "method %s invalid arg %s", method.Name, parg.Name)
				continue // keep going to catch more errors
			}
		}
		arg := &Field{NamePos(parg.NamePos), compileType(parg.Type, file, id.env)}
		args = append(args, arg)
	}
	return
}

func (id ifaceDefiner) defineStreamType(ptype parse.Type, file *File) *vdl.Type {
	if ptype == nil {
		return nil
	}
	if tn, ok := ptype.(*parse.TypeNamed); ok && tn.Name == "_" {
		// Special-case the _ placeholder, which means there's no stream type.
		return nil
	}
	return compileType(ptype, file, id.env)
}

func (id ifaceDefiner) defineTags(ptags []parse.ConstExpr, file *File) (tags []*vdl.Value) {
	for _, ptag := range ptags {
		if tag := compileConst("tag", nil, ptag, file, id.env); tag != nil {
			tags = append(tags, tag)
		}
	}
	return
}
