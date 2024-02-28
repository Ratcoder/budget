module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http exposing (..)
import Json.Decode
import Json.Encode
import Platform.Cmd as Cmd
import Task
import Time


main : Program () Model Msg
main =
    Browser.document
        { init = init
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }


init : () -> ( Model, Cmd Msg )
init _ =
    ( Stranger
        { usernameField = ""
        , passwordField = ""
        , responce = Unsent
        , isRegistering = False
        }
    , Cmd.none
    )


type Model
    = Stranger
        { usernameField : String
        , passwordField : String
        , responce : Response () String
        , isRegistering : Bool
        }
    | User
        { transactions : List Transaction
        , accounts : List Account
        , categories : List Category
        , dragedTransaction : Maybe Transaction
        , categoryNameField : String
        , categoryAvailableField : String
        , categoryBudgetedField : String
        , page : Page
        , date : String
        }


type Page
    = Budget
    | Accounts
    | AddTransaction
        { dateField : String
        , amountField : String
        , descriptionField : String
        , categoryIdField : String
        , isTransferField : Bool
        , error : Maybe String
        }


type alias Transaction =
    { id : Int
    , amount : Int
    , description : String
    , date : String
    , categoryId : Int
    , isTransfer : Bool
    }


type alias Account =
    { id : Int
    , name : String
    , balance : Int
    }


type alias Category =
    { id : Int
    , name : String
    , available : Int
    , budgeted : Int
    }


type Response success failure
    = Unsent
    | Loading
    | Success success
    | Failure failure


type Msg
    = ChangeUsername String
    | ChangePassword String
    | LoginSubmit
    | LoginResponse (Result Http.Error String)
    | DragEnterTransaction Transaction
    | DropTransaction Int
    | GetTransactions
    | GetTransactionsResponse (Result Http.Error (List Transaction))
    | GetCategories
    | GetCategoriesResponse (Result Http.Error (List Category))
    | GetAccounts
    | GetAccountsResponse (Result Http.Error (List Account))
    | UpdateCategory Int Category
    | ChangeCategoryName String
    | ChangeCategoryAvailable String
    | ChangeCategoryBudgeted String
    | AddCategoryResponse (Result Http.Error String)
    | AddCategory
    | SetPage Page
    | GetTime Time.Posix
    | ChangeTransactionDate String
    | ChangeTransactionAmount String
    | ChangeTransactionDescription String
    | ChangeTransactionCategoryId String
    | ChangeTransactionIsTransfer
    | SubmitTransaction
    | SubmitTransactionResponse (Result Http.Error String)
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case ( msg, model ) of
        ( ChangeUsername username, Stranger stranger ) ->
            ( Stranger { stranger | usernameField = username }, Cmd.none )

        ( ChangePassword password, Stranger stranger ) ->
            ( Stranger { stranger | passwordField = password }, Cmd.none )

        ( LoginSubmit, Stranger stranger ) ->
            ( Stranger { stranger | responce = Loading }
            , Http.post
                { url = "/api/login"
                , expect = Http.expectString LoginResponse
                , body =
                    Http.jsonBody <|
                        Json.Encode.object
                            [ ( "username", Json.Encode.string stranger.usernameField )
                            , ( "password", Json.Encode.string stranger.passwordField )
                            ]
                }
            )

        ( LoginResponse (Ok _), Stranger _ ) ->
            ( User
                { transactions = []
                , accounts = []
                , categories = []
                , dragedTransaction = Nothing
                , categoryNameField = ""
                , categoryAvailableField = ""
                , categoryBudgetedField = ""
                , page = Budget
                , date = ""
                }
            , Task.perform GetTime Time.now
            )
                |> batchUpdate [ GetTransactions, GetCategories, GetAccounts ]

        ( GetTime time, User user ) ->
            ( User { user | date = posixToDate Time.utc time }, Cmd.none )

        ( DragEnterTransaction transaction, User user ) ->
            ( User { user | dragedTransaction = Just transaction }, Cmd.none )

        ( DropTransaction categoryId, User user ) ->
            case user.dragedTransaction of
                Just transaction ->
                    if transaction.categoryId == categoryId then
                        ( model
                        , Cmd.none
                        )

                    else
                        let
                            categories =
                                List.map
                                    (\c ->
                                        if c.id == categoryId then
                                            { c | available = c.available + transaction.amount }

                                        else if c.id == transaction.categoryId then
                                            { c | available = c.available - transaction.amount }

                                        else
                                            c
                                    )
                                    user.categories
                        in
                        ( User
                            { user
                                | dragedTransaction = Nothing
                                , transactions =
                                    List.map
                                        (\t ->
                                            if t.id == transaction.id then
                                                { t | categoryId = categoryId }

                                            else
                                                t
                                        )
                                        user.transactions
                                , categories = categories
                            }
                        , Cmd.batch
                            (patch
                                { url = "/api/transactions"
                                , body =
                                    Http.jsonBody <|
                                        encodeTransaction
                                            { transaction | categoryId = categoryId }
                                , expect = Http.expectString (\_ -> NoOp)
                                }
                                :: List.filterMap
                                    (\c ->
                                        if c.id == categoryId || c.id == transaction.categoryId then
                                            Just
                                                (patch
                                                    { url = "/api/categories"
                                                    , body =
                                                        Http.jsonBody <|
                                                            encodeCategory c
                                                    , expect = Http.expectString (\_ -> NoOp)
                                                    }
                                                )

                                        else
                                            Nothing
                                    )
                                    categories
                            )
                        )

                Nothing ->
                    ( model
                    , Cmd.none
                    )

        ( GetTransactions, User _ ) ->
            ( model
            , Http.get
                { url = "/api/transactions"
                , expect = Http.expectJson GetTransactionsResponse (Json.Decode.list transactionDecoder)
                }
            )

        ( GetTransactionsResponse (Ok transactions), User user ) ->
            ( User { user | transactions = transactions }
            , Cmd.none
            )

        ( GetCategories, User _ ) ->
            ( model
            , Http.get
                { url = "/api/categories"
                , expect = Http.expectJson GetCategoriesResponse (Json.Decode.list categoryDecoder)
                }
            )

        ( GetCategoriesResponse (Ok categories), User user ) ->
            ( User { user | categories = categories }
            , Cmd.none
            )

        ( GetAccounts, User _ ) ->
            ( model
            , Http.get
                { url = "/api/accounts"
                , expect = Http.expectJson GetAccountsResponse (Json.Decode.list accountDecoder)
                }
            )

        ( GetAccountsResponse (Ok accounts), User user ) ->
            ( User { user | accounts = accounts }
            , Cmd.none
            )

        ( ChangeCategoryName name, User user ) ->
            ( User { user | categoryNameField = name }, Cmd.none )

        ( ChangeCategoryAvailable available, User user ) ->
            ( User { user | categoryAvailableField = available }, Cmd.none )

        ( ChangeCategoryBudgeted budgeted, User user ) ->
            ( User { user | categoryBudgetedField = budgeted }, Cmd.none )

        ( AddCategory, User user ) ->
            ( model
            , post
                { url = "/api/categories"
                , body =
                    Http.jsonBody <|
                        Json.Encode.object
                            [ ( "name", Json.Encode.string user.categoryNameField )
                            , ( "available", Json.Encode.int (String.toInt user.categoryAvailableField |> Maybe.withDefault 0) )
                            , ( "budgeted", Json.Encode.int (String.toInt user.categoryBudgetedField |> Maybe.withDefault 0) )
                            ]
                , expect = Http.expectString AddCategoryResponse
                }
            )

        ( AddCategoryResponse (Ok _), User _ ) ->
            update GetCategories model

        ( UpdateCategory id category, User user ) ->
            ( User
                { user
                    | categories =
                        List.map
                            (\c ->
                                if c.id == id then
                                    category

                                else
                                    c
                            )
                            user.categories
                }
            , patch
                { url = "/api/categories"
                , body =
                    Http.jsonBody <|
                        encodeCategory category
                , expect = Http.expectString (\_ -> NoOp)
                }
            )

        ( SetPage page, User user ) ->
            ( User { user | page = page }, Cmd.none )

        ( ChangeTransactionDate date, User user ) ->
            case user.page of
                AddTransaction page ->
                    ( User { user | page = AddTransaction { page | dateField = date } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        ( ChangeTransactionAmount amount, User user ) ->
            case user.page of
                AddTransaction page ->
                    ( User { user | page = AddTransaction { page | amountField = amount } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        ( ChangeTransactionDescription description, User user ) ->
            case user.page of
                AddTransaction page ->
                    ( User { user | page = AddTransaction { page | descriptionField = description } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        ( ChangeTransactionCategoryId categoryId, User user ) ->
            case user.page of
                AddTransaction page ->
                    ( User { user | page = AddTransaction { page | categoryIdField = categoryId } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        ( ChangeTransactionIsTransfer, User user ) ->
            case user.page of
                AddTransaction page ->
                    ( User { user | page = AddTransaction { page | isTransferField = not page.isTransferField } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        ( SubmitTransaction, User user ) ->
            case user.page of
                AddTransaction page ->
                    let
                        transaction =
                            Ok (Transaction 0 0 "" "" 0 False)
                                |> Result.andThen
                                    (\t ->
                                        case String.toInt page.amountField of
                                            Just a ->
                                                if a == 0 then
                                                    Err "Amount cannot be 0"

                                                else
                                                    Ok { t | amount = a }

                                            Nothing ->
                                                Err "Amount must be a number"
                                    )
                                |> Result.andThen
                                    (\t ->
                                        case String.toInt page.categoryIdField of
                                            Just c ->
                                                if c == 0 then
                                                    Err "Category field is required"

                                                else
                                                    Ok { t | categoryId = c }

                                            Nothing ->
                                                Err "Category must be a number"
                                    )
                                |> Result.andThen
                                    (\t ->
                                        if page.dateField == "" then
                                            Err "Date field is required"

                                        else
                                            Ok { t | date = page.dateField }
                                    )
                                |> Result.andThen
                                    (\t ->
                                        if page.descriptionField == "" then
                                            Err "Desciption field is required"

                                        else
                                            Ok { t | description = page.descriptionField }
                                    )
                                |> Result.andThen
                                    (\t ->
                                        Ok { t | isTransfer = page.isTransferField }
                                    )
                    in
                    case transaction of
                        Ok t ->
                            ( User { user | page = AddTransaction { page | error = Nothing } }
                            , post
                                { url = "/api/transactions"
                                , body =
                                    Http.jsonBody <|
                                        encodeTransaction t
                                , expect = Http.expectString SubmitTransactionResponse
                                }
                            )

                        Err e ->
                            ( User { user | page = AddTransaction { page | error = Just e } }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        _ ->
            ( model, Cmd.none )


batchUpdate : List Msg -> ( Model, Cmd Msg ) -> ( Model, Cmd Msg )
batchUpdate msgs ( model, cmd ) =
    let
        ( finalModel, finalCmd ) =
            List.foldl
                (\msg ( accModel, accCmd ) ->
                    let
                        ( newModel, newCmd ) =
                            update msg accModel
                    in
                    ( newModel, Cmd.batch [ accCmd, newCmd ] )
                )
                ( model, cmd )
                msgs
    in
    ( finalModel, finalCmd )


view : Model -> Browser.Document Msg
view model =
    case model of
        Stranger stranger ->
            { title = "Budget"
            , body =
                [ h2 [] [ text "Budget" ]
                , div []
                    [ button [ onClick LoginSubmit ] [ text "Login" ]
                    , input [ type_ "text", value stranger.usernameField, onInput ChangeUsername ] []
                    , input [ type_ "password", value stranger.passwordField, onInput ChangePassword ] []
                    , p []
                        [ case stranger.responce of
                            Unsent ->
                                text ""

                            Loading ->
                                text "Loading..."

                            Success _ ->
                                text "Success!"

                            Failure err ->
                                text <| "Failure: " ++ err
                        ]
                    ]
                ]
            }

        User user ->
            { title = "Budget"
            , body =
                [ div []
                    [ button [ onClick (SetPage Budget) ] [ text "Budget" ]
                    , button [ onClick (SetPage Accounts) ] [ text "Accounts" ]
                    , button [ onClick (SetPage (AddTransaction { dateField = user.date, amountField = "", descriptionField = "", categoryIdField = "", isTransferField = False, error = Nothing })) ] [ text "Add Transaction" ]
                    ]
                , case user.page of
                    Budget ->
                        div []
                            [ h2 [] [ text "Budget" ]
                            , button [ onClick GetTransactions ] [ text "Get Transactions" ]
                            , button [ onClick GetCategories ] [ text "Get Categories" ]
                            , button [ onClick GetAccounts ] [ text "Get Accounts" ]
                            , h2 [] [ text <| "Uncategorized: " ++ formatDollars (List.foldl (\a sum -> sum + a.balance) 0 user.accounts - List.foldl (\c sum -> sum + c.available) 0 user.categories) ]
                            , ul [] <| List.map (\t -> li [] [ viewTransaction t ]) <| List.filter (\t -> t.categoryId == 0 && not t.isTransfer) <| List.filter (\t -> t.date >= String.dropRight 2 user.date ++ "01") user.transactions
                            , h2 [] [ text "Categories:" ]
                            , ul [ class "category-list" ] <|
                                List.map
                                    (\c ->
                                        li [ class "category", preventDefaultOn "drop" (Json.Decode.succeed ( DropTransaction c.id, True )), preventDefaultOn "dragover" (Json.Decode.succeed ( NoOp, True )) ]
                                            [ details []
                                                [ summary []
                                                    [ div [ style "display" "flex", style "gap" "3ch" ]
                                                        [ div [ style "flex" "1" ] [ text c.name ]
                                                        , div [ style "text-align" "right", style "width" "15ch" ] [ input [ type_ "number", value (String.fromFloat (toFloat c.available / 100)), onInput (\s -> UpdateCategory c.id { c | available = round <| 100 * (String.toFloat s |> Maybe.withDefault 0) }) ] [] ]
                                                        , div [ style "text-align" "right", style "width" "15ch" ] [ input [ type_ "number", value (String.fromFloat (toFloat c.budgeted / 100)), onInput (\s -> UpdateCategory c.id { c | budgeted = round <| 100 * (String.toFloat s |> Maybe.withDefault 0) }) ] [] ]
                                                        ]
                                                    ]
                                                , ul [ class "category-transaction-list" ] <| List.map (\t -> li [] [ viewTransaction t ]) <| List.sortBy .date <| List.filter (\t -> t.categoryId == c.id && t.date >= String.dropRight 2 user.date ++ "01") user.transactions
                                                ]
                                            ]
                                    )
                                    user.categories
                            , div []
                                [ input [ type_ "text", value user.categoryNameField, onInput ChangeCategoryName ] []
                                , input [ type_ "number", value user.categoryAvailableField, onInput ChangeCategoryAvailable ] []
                                , input [ type_ "number", value user.categoryBudgetedField, onInput ChangeCategoryBudgeted ] []
                                , button [ onClick AddCategory ] [ text "Add Category" ]
                                ]
                            ]

                    AddTransaction page ->
                        div []
                            [ h2 [] [ text "Add Transaction" ]
                            , label [ for "date" ] [ text "Date" ]
                            , input [ type_ "date", id "date", value page.dateField, onInput ChangeTransactionDate ] []
                            , label [ for "amount" ] [ text "Amount" ]
                            , input [ type_ "number", id "amount", value page.amountField, onInput ChangeTransactionAmount ] []
                            , label [ for "description" ] [ text "Description" ]
                            , input [ type_ "text", id "description", value page.descriptionField, onInput ChangeTransactionDescription ] []
                            , label [ for "category" ] [ text "Category" ]
                            , select [ id "category", value page.categoryIdField, onInput ChangeTransactionCategoryId ]
                                (option [ value "" ] [ text "" ]
                                    :: List.map (\c -> option [ value <| String.fromInt c.id ] [ text c.name ]) user.categories
                                )
                            , label [ for "isTransfer" ] [ text "Is Transfer" ]
                            , input [ type_ "checkbox", id "isTransfer", checked page.isTransferField, onClick ChangeTransactionIsTransfer ] []
                            , button [ onClick SubmitTransaction ] [ text "Submit" ]
                            , p []
                                [ case page.error of
                                    Just e ->
                                        text e

                                    Nothing ->
                                        text ""
                                ]
                            ]

                    Accounts ->
                        div []
                            [ h2 [] [ text "Accounts" ]
                            , button [ onClick GetAccounts ] [ text "Get Accounts" ]
                            , ul [] <| List.map (\a -> li [] [ text <| a.name ++ " - " ++ formatDollars a.balance ]) user.accounts
                            ]
                ]
            }


viewTransaction : Transaction -> Html Msg
viewTransaction transaction =
    div [ class "transaction", draggable "true", on "dragstart" (Json.Decode.succeed (DragEnterTransaction transaction)) ]
        [ div []
            [ span [ class "transaction-description" ] [ text <| transaction.description ]
            , span [ class "transaction-date" ] [ text <| formatDate transaction.date ]
            ]
        , div []
            [ span [] [ text <| formatDollars transaction.amount ]
            ]
        ]


posixToDate : Time.Zone -> Time.Posix -> String
posixToDate zone posix =
    String.fromInt (Time.toYear zone posix) ++ "-" ++ String.padLeft 2 '0' (String.fromInt (monthToInt (Time.toMonth zone posix))) ++ "-" ++ String.padLeft 2 '0' (String.fromInt (Time.toDay zone posix))


monthToInt : Time.Month -> Int
monthToInt month =
    case month of
        Time.Jan ->
            1

        Time.Feb ->
            2

        Time.Mar ->
            3

        Time.Apr ->
            4

        Time.May ->
            5

        Time.Jun ->
            6

        Time.Jul ->
            7

        Time.Aug ->
            8

        Time.Sep ->
            9

        Time.Oct ->
            10

        Time.Nov ->
            11

        Time.Dec ->
            12


formatDate : String -> String
formatDate date =
    case String.split "-" date of
        [ year, month, day ] ->
            String.left 3 (String.toInt month |> Maybe.andThen monthToName |> Maybe.withDefault month) ++ " " ++ String.padLeft 2 '0' day ++ ", " ++ year

        _ ->
            date


monthToName : Int -> Maybe String
monthToName month =
    case month of
        1 ->
            Just "January"

        2 ->
            Just "February"

        3 ->
            Just "March"

        4 ->
            Just "April"

        5 ->
            Just "May"

        6 ->
            Just "June"

        7 ->
            Just "July"

        8 ->
            Just "August"

        9 ->
            Just "September"

        10 ->
            Just "October"

        11 ->
            Just "November"

        12 ->
            Just "December"

        _ ->
            Nothing


formatDollars : Int -> String
formatDollars amount =
    (if amount >= 0 then
        ""

     else
        "-"
    )
        ++ "$"
        ++ String.fromInt (abs amount // 100)
        ++ "."
        ++ String.padLeft 2 '0' (String.fromInt <| abs amount - (abs amount // 100) * 100)


transactionDecoder : Json.Decode.Decoder Transaction
transactionDecoder =
    Json.Decode.map6 Transaction
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "amount" Json.Decode.int)
        (Json.Decode.field "description" Json.Decode.string)
        (Json.Decode.field "date" Json.Decode.string)
        (Json.Decode.field "category_id" Json.Decode.int)
        (Json.Decode.field "is_transfer" Json.Decode.bool)


encodeTransaction : Transaction -> Json.Encode.Value
encodeTransaction transaction =
    Json.Encode.object
        [ ( "id", Json.Encode.int transaction.id )
        , ( "amount", Json.Encode.int transaction.amount )
        , ( "description", Json.Encode.string transaction.description )
        , ( "date", Json.Encode.string transaction.date )
        , ( "category_id", Json.Encode.int transaction.categoryId )
        , ( "is_transfer", Json.Encode.bool transaction.isTransfer )
        ]


categoryDecoder : Json.Decode.Decoder Category
categoryDecoder =
    Json.Decode.map4 Category
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "name" Json.Decode.string)
        (Json.Decode.field "available" Json.Decode.int)
        (Json.Decode.field "budgeted" Json.Decode.int)


encodeCategory : Category -> Json.Encode.Value
encodeCategory category =
    Json.Encode.object
        [ ( "id", Json.Encode.int category.id )
        , ( "name", Json.Encode.string category.name )
        , ( "available", Json.Encode.int category.available )
        , ( "budgeted", Json.Encode.int category.budgeted )
        ]


accountDecoder : Json.Decode.Decoder Account
accountDecoder =
    Json.Decode.map3 Account
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "name" Json.Decode.string)
        (Json.Decode.field "balance" Json.Decode.int)


patch :
    { url : String
    , body : Http.Body
    , expect : Http.Expect msg
    }
    -> Cmd msg
patch p =
    Http.request
        { method = "PATCH"
        , headers = []
        , url = p.url
        , body = p.body
        , expect = p.expect
        , timeout = Nothing
        , tracker = Nothing
        }
