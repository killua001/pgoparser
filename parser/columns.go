package parser

import (
	"fmt"
	. "github.com/elliotcourant/pgoparser/keywords"
	"github.com/elliotcourant/pgoparser/symbols"
	"github.com/elliotcourant/pgoparser/tokens"
	"github.com/elliotcourant/pgoparser/tree"
	"github.com/elliotcourant/pgoparser/words"
)

func (p *parser) parseColumns() ([]tree.ColumnDefinition, []tree.Constraint, error) {
	// If we cant find the column opening paren then just return?
	if !p.consumeTokenMaybe(symbols.LeftParentheses) || p.consumeTokenMaybe(symbols.RightParentheses) {
		return nil, nil, nil
	}

	columns := make([]tree.ColumnDefinition, 0)
	constraints := make([]tree.Constraint, 0)

	for {
		// Handle constraints here.
		if constraint, err := p.parseOptionalTableConstraint(); err != nil {
			return nil, nil, err
		} else if constraint != nil {
			constraints = append(constraints, *constraint)
		} else if _, ok := p.peakToken().(words.Word); ok { // Try to parse a column name.
			column, columnConstraints, err := p.parseColumnDefinition()
			if err != nil {
				return nil, nil, err
			}

			columns = append(columns, column)
			constraints = append(constraints, columnConstraints...)
		} else {
			return nil, nil, p.expected("column name or constraint definition", p.peakToken())
		}

		comma := p.consumeTokenMaybe(symbols.Comma)
		if p.consumeTokenMaybe(symbols.RightParentheses) {
			break
		} else if !comma {
			return nil, nil, p.expected("',' or ')' after column definition", p.peakToken())
		}
	}

	return columns, constraints, nil
}

func (p *parser) parseOptionalTableConstraint() (constraint *tree.Constraint, err error) {
	var name *string

	// When parsing a constraint, if we grab the constraint keyword first, then the constraint must be named and is
	// specifically declared. If we fail to parse the name or the constraint now then we want to return an error. If not
	// we can still fail gracefully with no error.
	if ok := p.parseKeyword(CONSTRAINT); ok {
		name, err = p.parseConstraintName()
		if err != nil {
			return nil, err
		}
	}

	token := p.nextToken()
	switch token {
	case NOT:
		if !p.parseKeyword(NULL) {
			return nil, p.expected("NULL", p.peakToken())
		}

		constraint = &tree.Constraint{
			Type:              tree.ConstraintTypeNotNull,
			Name:              name,
			Deferrable:        false,
			InitiallyDeferred: false,
			IsNoInherit:       false,
			RawExpression:     nil,
			CookedExpression:  nil,
		}
	case NULL:
	case PRIMARY, UNIQUE:
	case FOREIGN:
		if !p.parseKeyword(KEY) {
			return nil, p.expected("KEY", p.peakToken())
		}

		panic("foreign key not implemented")

		return constraint, nil
	case CHECK:
	default:
		if name != nil {
			return nil, p.expected("PRIMARY, UNIQUE, FOREIGN or CHECK", token)
		} else {
			p.previousToken() // Move the cursor back, we don't need to parse this constraint.
			return nil, nil
		}
	}

	switch {
	case p.parseKeywords(NOT, DEFERRABLE):
		constraint.Deferrable = false
	case p.parseKeywords(DEFERRABLE):
		constraint.Deferrable = true
	}

	switch {
	case p.parseKeywords(INITIALLY, DEFERRED):
		constraint.InitiallyDeferred = true
	case p.parseKeywords(INITIALLY, IMMEDIATE):
		constraint.InitiallyDeferred = false
	}

	return constraint, nil
}

func (p *parser) parseColumnDefinition() (definition tree.ColumnDefinition, constraints []tree.Constraint, err error) {
	columnName, err := p.parseIdentifier()
	if err != nil {
		return definition, nil, err
	}

	definition.ColumnName = columnName.String()

	definition.Type, err = p.parseDataType()
	if err != nil {
		return definition, nil, err
	}

	// If there is a collation defined, parse it.
	if p.parseKeyword(COLLATE) {
		collation, err := p.parseObjectName()
		if err != nil {
			return definition, nil, err
		}

		// TODO (elliotcourant) Do something with the collation
		fmt.Sprint(collation)
	}

	// definition.Options = make([]tree.ColumnOption, 0)
OptionLoop:
	for {
		nextToken := p.peakToken()
		switch nextToken {
		case tokens.EOF{}, symbols.Comma, symbols.RightParentheses:
			break OptionLoop
		default:
			option, err := p.parseColumnOptionDefinition()
			if err != nil {
				return definition, nil, err
			}

			switch option.(type) {
			// If the column has a primary key or not null option then set nullable to false.
			case tree.PrimaryKey, tree.NotNull:
				definition.IsNotNull = true
			}

			// definition.Options = append(definition.Options, option)
		}
	}

	return definition, nil, nil
}

func (p *parser) parseColumnOptionDefinition() (_ tree.ColumnOption, err error) {
	var name tokens.Token
	if p.parseKeyword(CONSTRAINT) {
		name, err = p.parseIdentifier()
		if err != nil {
			return nil, err
		}

		// TODO (elliotcourant) do something with name.
		fmt.Sprint(name)
	}

	nextToken := p.nextToken()
	switch nextToken {
	case NOT:
		if !p.parseKeyword(NULL) {
			return nil, p.expected("NULL", p.peakToken())
		}

		return tree.NotNull(0), nil
	case NULL:
		return tree.Null(0), nil
	case DEFAULT:
	case PRIMARY:
		if !p.parseKeyword(KEY) {
			return nil, p.expected("KEY", p.peakToken())
		}

		return tree.PrimaryKey(0), nil
	case UNIQUE:
		return tree.Unique(0), nil
	case REFERENCES:
		foreignTable, err := p.parseTableName()
		if err != nil {
			return nil, err
		}

		referredColumns, err := p.parseParenthesizedColumnList(true)
		if err != nil {
			return nil, err
		}

		foreignKey := tree.ForeignKey{
			ForeignTable:    foreignTable,
			ReferredColumns: make([]tree.ColumnName, len(referredColumns)),
			OnDelete:        0,
			OnUpdate:        0,
		}

		for i, column := range referredColumns {
			foreignKey.ReferredColumns[i], err = tree.NewColumnName(tokens.NewObjectName(column))
			if err != nil {
				return nil, err
			}
		}

		for {
			// If on delete has not yet been set and the next two keywords are ON DELETE, then parse the delete.
			if foreignKey.OnDelete == 0 && p.parseKeywords(ON, DELETE) {
				foreignKey.OnDelete, err = p.parseReferentialAction()
				if err != nil {
					return nil, err
				}
			} else if foreignKey.OnUpdate == 0 && p.parseKeywords(ON, UPDATE) {
				foreignKey.OnUpdate, err = p.parseReferentialAction()
				if err != nil {
					return nil, err
				}
			} else {
				break
			}
		}

		return foreignKey, nil
	case CHECK:

	default:
		return nil, p.expected("column option", nextToken)
	}

	return nil, nil
}

func (p *parser) parseParenthesizedColumnList(optional bool) ([]tokens.Token, error) {
	if p.consumeTokenMaybe(symbols.LeftParentheses) {
		columns, err := p.parseCommaSeparated(p.parseIdentifier)
		if err != nil {
			return nil, err
		}

		return columns, p.expectToken(symbols.RightParentheses)
	} else if optional {
		return nil, nil
	} else {
		return nil, p.expected("a list of columns in parentheses", p.peakToken())
	}
}

func (p *parser) parseReferentialAction() (tree.ReferenceAction, error) {
	nextToken := p.nextToken()
	switch nextToken {
	case RESTRICT:
		return tree.Restrict, nil
	case CASCADE:
		return tree.Cascade, nil
	case SET:
		switch p.peakToken() {
		case NULL:
			return tree.SetNull, nil
		case DEFAULT:
			return tree.SetDefault, nil
		default:
			return 0, p.expected("NULL or DEFAULT", p.peakToken())
		}
	case NO:
		if !p.parseKeyword(ACTION) {
			return 0, p.expected("ACTION", p.peakToken())
		}
		return tree.NoAction, nil
	default:
		return 0, p.expected("one of RESTRICT, CASCADE, SET NULL, NO ACTION or SET DEFAULT", nextToken)
	}
}
