package parser

import (
	"fmt"
)

// Abstract, should not get called but needed to cast abstract struct to Expression
func (e *Positioned) Label() string { return "Positioned" }

// Concrete
func (e *StepExpression) Label() string              { return "Step" }
func (e *AccessExpression) Label() string            { return "'[]' expression" }
func (e *AndExpression) Label() string               { return "'and' expression" }
func (e *ArithmeticExpression) Label() string        { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *Application) Label() string                 { return "Application" }
func (e *AssignmentExpression) Label() string        { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *AttributeOperation) Label() string          { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *AttributesOperation) Label() string         { return "AttributesOperation" }
func (e *BlockExpression) Label() string             { return "Block Expression" }
func (e *CallMethodExpression) Label() string        { return "Method Call" }
func (e *CallNamedFunctionExpression) Label() string { return "Function Call" }
func (e *CapabilityMapping) Label() string           { return "Capability Mapping" }
func (e *CaseExpression) Label() string              { return "'case' statement" }
func (e *CaseOption) Label() string                  { return "CaseOption" }
func (e *CollectExpression) Label() string           { return "CollectExpression" }
func (e *ComparisonExpression) Label() string        { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *ConcatenatedString) Label() string          { return "Concatenated String" }
func (e *EppExpression) Label() string               { return "Epp Template" }
func (e *ExportedQuery) Label() string               { return "Exported Query" }
func (e *FunctionDefinition) Label() string          { return "Function Definition" }
func (e *HeredocExpression) Label() string           { return "Heredoc" }
func (e *HostClassDefinition) Label() string         { return "Host Class Definition" }
func (e *IfExpression) Label() string                { return "'if' statement" }
func (e *InExpression) Label() string                { return "'in' expression" }
func (e *KeyedEntry) Label() string                  { return "Hash Entry" }
func (e *LiteralBoolean) Label() string              { return "Literal Boolean" }
func (e *LiteralDefault) Label() string              { return "'default' expression" }
func (e *LiteralFloat) Label() string                { return "Literal Float" }
func (e *LiteralHash) Label() string                 { return "Hash Expression" }
func (e *LiteralInteger) Label() string              { return "Literal Integer" }
func (e *LiteralList) Label() string                 { return "Array expression" }
func (e *LiteralString) Label() string               { return "Literal String" }
func (e *LiteralUndef) Label() string                { return "'undef' expression" }
func (e *Locator) Label() string                     { return "Locator" }
func (e *MatchExpression) Label() string             { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *NamedAccessExpression) Label() string       { return "'.' expression" }
func (e *NodeDefinition) Label() string              { return "Node Definition" }
func (e *Nop) Label() string                         { return "Nop" }
func (e *NotExpression) Label() string               { return "'!' expression" }
func (e *OrExpression) Label() string                { return "'or' expression" }
func (e *Parameter) Label() string                   { return "Parameter Definition" }
func (e *Program) Label() string                     { return "Program" }
func (e *QualifiedName) Label() string               { return "Name" }
func (e *QualifiedReference) Label() string          { return "Type-Name" }
func (e *RelationshipExpression) Label() string      { return fmt.Sprintf("'%s' expression", e.operator) }
func (e *RenderExpression) Label() string            { return "Epp Interpolated Expression" }
func (e *RenderStringExpression) Label() string      { return "Epp Text" }
func (e *RegexpExpression) Label() string            { return "Regular Expression" }
func (e *ReservedWord) Label() string                { return fmt.Sprintf("Reserved Word '%s'", e.word) }
func (e *ResourceBody) Label() string                { return "Resource Instance Definition" }
func (e *ResourceDefaultsExpression) Label() string  { return "Resource Defaults Expression" }
func (e *ResourceExpression) Label() string          { return "Resource Statement" }
func (e *ResourceOverrideExpression) Label() string  { return "Resource Override" }
func (e *ResourceTypeDefinition) Label() string      { return "'define' expression" }
func (e *SelectorEntry) Label() string               { return "Selector option" }
func (e *SelectorExpression) Label() string          { return "Selector expression" }
func (e *SiteDefinition) Label() string              { return "Site Definition" }
func (e *TextExpression) Label() string              { return "Text expression" }
func (e *TypeAlias) Label() string                   { return "Type Alias" }
func (e *TypeDefinition) Label() string              { return "Type Definition" }
func (e *TypeMapping) Label() string                 { return "Type Mapping" }
func (e *UnaryMinusExpression) Label() string        { return "Unary Minus" }
func (e *UnfoldExpression) Label() string            { return "Unfold" }
func (e *UnlessExpression) Label() string            { return "'unless' statement" }
func (e *VariableExpression) Label() string          { return "Variable" }
func (e *VirtualQuery) Label() string                { return "Virtual Query" }
